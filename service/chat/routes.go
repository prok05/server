package chat

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/cache"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/service/lesson"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	store        types.ChatStore
	userStore    types.UserStore
	messageStore types.MessageStore
	tokenCache   *cache.TokenCache
}

func NewHandler(store types.ChatStore, userStore types.UserStore, messageStore types.MessageStore, tokenCache *cache.TokenCache) *Handler {
	return &Handler{
		store:        store,
		tokenCache:   tokenCache,
		userStore:    userStore,
		messageStore: messageStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/chats", h.CreateChat).Methods("POST")
	router.HandleFunc("/chats", h.GetAllChats).Methods("GET")
	router.HandleFunc("/chats/{chatID}", h.GetChatByID).Methods("GET")
	router.HandleFunc("/chats/get/{userID}", h.GetChatByIDs).Methods("GET")
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

	parts := []int{1, 2, 3}

	if err := h.store.CreateChat(&chat, parts); err != nil {
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

	role, ok := claims["role"].(string)
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

	if role == "student" {
		token, err := h.tokenCache.GetToken()
		if err != nil {
			log.Println("no alpha token")
			return
		}

		teachersIds, err := lesson.GetStudentTeachersIDs(userID, 3, 0, token)
		if err != nil {
			log.Println(err)
		}

		teachers, err := h.userStore.FindUsersByIDs(teachersIds)
		if err != nil {
			log.Printf("teachers not found: %v", err)
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("teachers not found"))
			return
		}

		utils.WriteJSON(w, http.StatusOK, teachers)
	} else if role == "teacher" {
		chats, err := h.store.GetAllChats(userID)
		if err != nil {
			log.Println("error getting chats: ", err)
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		utils.WriteJSON(w, http.StatusOK, chats)
	}

	//
	//var chats []types.AllChatsItem
	//
	//chats, err = h.store.GetAllChats(userID)
	//if err != nil {
	//	utils.WriteError(w, http.StatusInternalServerError, err)
	//	return
	//}
	//
	//resp := types.AllChatsResponse{
	//	Count: len(chats),
	//	Items: chats,
	//}
	//
	//if len(resp.Items) == 0 {
	//	resp.Items = []types.AllChatsItem{}
	//}
	//

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

func (h *Handler) GetChatByIDs(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	userIDParam := vars["userID"]
	userIDParamInt, err := strconv.Atoi(userIDParam)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid chat id"))
		return
	}

	chat, err := h.store.GetChatByUserIDs(userID, userIDParamInt)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("chat not found"))
		log.Println(err)
		return
	}

	fmt.Println(chat)

	if chat == nil {
		utils.WriteJSON(w, http.StatusOK, chat)
	} else {
		messages, err := h.messageStore.GetMessages(chat.ID, 30, 0)
		if err != nil {
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("messages not found"))
			log.Println(err)
			return
		}
		utils.WriteJSON(w, http.StatusOK, messages)
	}
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
