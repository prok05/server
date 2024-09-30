package ws

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/types"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
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

func ServeWS(hub *Hub, store types.MessageStore, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка при апгрейде соединения:", err)
		return
	}

	client := NewClient(hub, conn)
	hub.register <- client

	// Запуск процессов чтения и записи для WebSocket клиента
	go client.writePump()
	go client.readPump(hub, store) // Передаем store
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: userID=%d", client.userID)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: userID=%d", client.userID)
			}
			h.mu.Unlock()
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

		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userIDStr, ok := claims["userID"].(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("Unable to convert userID: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Websocket upgrade error:", err)
			return
		}

		client := &Client{
			conn:    conn,
			send:    make(chan types.Message),
			userID:  userID,
			chatIDs: make(map[int]bool),
		}

		hub.register <- client

		go client.writePump()
		go client.readPump(hub, store)
	}
}

func (c *Client) readPump(hub *Hub, store types.MessageStore) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg types.Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Websocket read error: %v", err)
			}
			break
		}

		isMember, err := store.IsUserInChat(msg.ChatID, c.userID)
		if err != nil || !isMember {
			c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "You are not a member of this chat"))
			return
		}

		message := &types.Message{
			ChatID:    msg.ChatID,
			SenderID:  c.userID,
			Content:   msg.Content,
			CreatedAt: time.Now(),
		}

		if err := store.SaveMessage(message); err != nil {
			c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to save message"))
			return
		}

		outgoing := types.Message{
			ID:        message.ID,
			ChatID:    message.ChatID,
			SenderID:  message.SenderID,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
		}

		hub.broadcast <- outgoing
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		msg, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.conn.WriteJSON(msg)
		if err != nil {
			log.Println("Websocket write error:", err)
			return
		}
	}
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		conn:    conn,
		send:    make(chan types.Message),
		chatIDs: make(map[int]bool),
	}
}
