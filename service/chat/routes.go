package chat

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	store types.ChatStore
}

func NewHandler(store types.ChatStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/chats", h.CreateChat).Methods("POST")
	router.HandleFunc("/chats", h.GetAllChats).Methods("GET")
	router.HandleFunc("/chats/{chatID}", h.GetChatByID).Methods("GET")
	router.HandleFunc("/chats/{chatID}", h.DeleteChat).Methods("DELETE")
}

func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
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

	var chat types.Chat
	if err := utils.ParseJSON(r, &chat); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.store.CreateChat(&chat); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, chat)
}

func (h *Handler) GetAllChats(w http.ResponseWriter, r *http.Request) {
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

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("no token"))
		return
	}

	userIDStr, ok := claims["userID"].(string)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Unable to convert userID: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var chats []types.Chat

	chats, err = h.store.GetAllChats(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := types.AllChatsResponse{
		Count: len(chats),
		Items: chats,
	}

	if len(resp.Items) == 0 {
		resp.Items = []types.Chat{}
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetChatByID(w http.ResponseWriter, r *http.Request) {
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
	chatID := vars["chatID"]
	chatIDInt, err := strconv.Atoi(chatID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid chat id"))
		return
	}

	chat, err := h.store.GetChatByID(chatIDInt)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("chat not found"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, chat)

}

func (h *Handler) DeleteChat(w http.ResponseWriter, r *http.Request) {
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
	chatID := vars["chatID"]
	chatIDInt, err := strconv.Atoi(chatID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid chat id"))
		return
	}

	if err := h.store.DeleteChat(chatIDInt); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusNoContent, nil)
}
