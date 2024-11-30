package message

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/cache"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	store      types.MessageStore
	chatStore  types.ChatStore
	tokenCache *cache.TokenCache
}

func NewHandler(store types.MessageStore, chatStore types.ChatStore, tokenCache *cache.TokenCache) *Handler {
	return &Handler{
		store:      store,
		chatStore:  chatStore,
		tokenCache: tokenCache,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/messages", h.SendMessage).Methods("POST")
	router.HandleFunc("/chats/{chatID}/messages", h.GetMessages).Methods("GET")
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
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

	role, ok := claims["role"].(string)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	var payload types.MessagePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if role == "teacher" {
		if err := h.store.SaveMessage(&payload.Message); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		utils.WriteJSON(w, http.StatusCreated, nil)
	} else if role == "student" {
		chat, err := h.chatStore.GetChatByUserIDs(userID, payload.UserID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		if chat == nil {
			chat := types.Chat{
				ChatType: "",
				Name:     "",
			}
			err = h.chatStore.CreateChat(&chat, []int{userID, payload.UserID})
			if err != nil {
				log.Println("unable to create new chat:", err)
				return
			}
		}

		payload.Message.ChatID = chat.ID
		payload.Message.SenderID = userID

		if err := h.store.SaveMessage(&payload.Message); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		utils.WriteJSON(w, http.StatusCreated, nil)
	}
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
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

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20 // Значение по умолчанию
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // Значение по умолчанию
	}

	messages, err := h.store.GetMessages(chatIDInt, limit, offset)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("cant save message"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, messages)
}
