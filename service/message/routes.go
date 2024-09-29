package message

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"net/http"
	"strconv"
)

type Handler struct {
	store types.MessageStore
}

func NewHandler(store types.MessageStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/messages", h.SendMessage).Methods("POST")
	router.HandleFunc("/chats/{chatID}/messages", h.GetMessages).Methods("POST")
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

	var message types.Message
	if err := utils.ParseJSON(r, &message); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.store.SaveMessage(&message); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, nil)
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
		limit = 10 // Значение по умолчанию
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
