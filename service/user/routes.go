package user

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/config"
	"github.com/prok05/ecom/service/alpha"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	store types.UserStore
}

func NewHandler(store types.UserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods(http.MethodPost)
	router.HandleFunc("/register", h.handleRegister).Methods(http.MethodPost)
	router.HandleFunc("/logout", h.handleLogout).Methods(http.MethodPost)

	router.HandleFunc("/alpha/users/{userID}", h.handleAlphaGetUser).Methods(http.MethodGet)
}

// Логин пользователя
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Опа")
	// парсинг payload
	var payload types.LoginUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	// валидация payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	u, err := h.store.FindUserByPhone(payload.Phone)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(payload.Password)) {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("invalid phone or password"))
		return
	}

	secret := []byte(config.Envs.JWTSecret)
	token, err := auth.CreateJWT(secret, u.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour * 30),
		HttpOnly: true,  // Куки недоступны для JS
		Secure:   false, // Используйте true при HTTPS
		Path:     "/",
	})

	//utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "login successful"})

}

// Регистрация пользователя
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// парсинг payload
	var payload types.RegisterUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	// валидация payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// проверка существует ли пользователь
	_, err := h.store.FindUserByPhone(payload.Phone)
	if err == nil {
		utils.WriteError(w, http.StatusConflict, fmt.Errorf("user with phone %s "+
			"already exists", payload.Phone))
		return
	}

	// получения токена AlphaCRM
	token, err := alpha.GetAlphaToken()
	if err != nil {
		fmt.Printf("Error getting alpha token: %v\n", err)
		return
	}

	// получение пользователя из Alpha
	alphaUser, err := alpha.GetUserIDByPhone(payload.Phone, payload.Role, token)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("Пользователь с таким номером телефона не найден: %s", payload.Phone))
		return
	}

	fullName := strings.Split(alphaUser.Name, " ")
	firstName := fullName[1]
	lastName := fullName[0]
	var middleName string
	if len(fullName) == 3 {
		middleName = fullName[2]
	} else {
		middleName = ""
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.CreateUser(types.User{
		ID:         alphaUser.ID,
		Phone:      payload.Phone,
		FirstName:  firstName,
		LastName:   lastName,
		MiddleName: middleName,
		Role:       payload.Role,
		Password:   hashedPassword,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, nil)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-1),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "logout successful"})

}

func (h *Handler) handleAlphaGetUser(w http.ResponseWriter, r *http.Request) {
	tokenString := auth.GetTokenFromRequest(r)
	if tokenString == "" {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("missing or invalid JWT token"))
		return
	}

	jwtToken, err := auth.ValidateToken(tokenString)
	if err != nil || !jwtToken.Valid {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	vars := mux.Vars(r)
	str, ok := vars["userID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing user ID"))
		return
	}

	userID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	token, err := alpha.GetAlphaToken()
	if err != nil {
		fmt.Printf("Error getting alpha token: %v\n", err)
		return
	}

	alphaUser, err := alpha.GetUserById(userID, token)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("no users with such user id: %s", userID))
		return
	}

	utils.WriteJSON(w, http.StatusOK, alphaUser)
}
