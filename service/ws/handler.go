package ws

import (
	"github.com/gorilla/websocket"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/types"
	"log"
	"net/http"
	"sync"
)

type Client struct {
	conn    *websocket.Conn
	send    chan types.Message
	userID  int
	chatIDs map[int]bool
	mu      sync.Mutex
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan types.Message
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan types.Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("client registered: userID=%d", client.userID)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: userID=%d", client.userID)
			}
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				if client.chatIDs[message.ChatID] {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Handler(hub *Hub, store types.MessageStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := auth.GetTokenFromRequest(r)
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		jwtToken, err := auth.ValidateToken(tokenString)
		if err != nil || !jwtToken.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := jwtToken.Claims.(map[string]interface{})["userID"].(float64)

	}
}
