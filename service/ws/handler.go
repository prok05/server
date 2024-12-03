package ws

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/prok05/ecom/cache"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
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
	role    string
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

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		conn:    conn,
		send:    make(chan types.Message),
		chatIDs: make(map[int]bool),
	}
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

func Handler(hub *Hub, messageStore types.MessageStore, chatStore types.ChatStore, userStore types.UserStore, tokenCache *cache.TokenCache) http.HandlerFunc {
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

		role, ok := claims["role"].(string)
		if !ok {
			utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
			log.Printf("Unable to get role: %v", err)
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
			role:    role,
			userID:  userID,
			chatIDs: make(map[int]bool),
		}

		if client.role == "student" {
			//token, err := tokenCache.GetToken()
			//if err != nil {
			//	log.Println("no alpha token")
			//	return
			//}
			//
			//teachersIds, err := lesson.GetStudentTeachersIDs(userID, 3, 0, token)
			//if err != nil {
			//	log.Println(err)
			//}
			//
			//teachers, err := userStore.FindUsersByIDs(teachersIds)
			//if err != nil {
			//	log.Printf("teachers not found: %v", err)
			//	utils.WriteError(w, http.StatusNotFound, fmt.Errorf("teachers not found"))
			//	return
			//}
			//
			//for _, teacher := range *teachers {
			//	client.chatIDs[teacher.ID] = true
			//}
			chats, err := chatStore.GetAllChats(userID)
			if err != nil {
				log.Printf("Error retrieving user chats: %v", err)
				http.Error(w, "Failed to retrieve chats", http.StatusInternalServerError)
				return
			}
			for _, chat := range chats {
				client.chatIDs[chat.ID] = true
			}
		} else if client.role == "teacher" {
			chats, err := chatStore.GetAllChats(userID)
			if err != nil {
				log.Printf("Error retrieving user chats: %v", err)
				http.Error(w, "Failed to retrieve chats", http.StatusInternalServerError)
				return
			}
			for _, chat := range chats {
				client.chatIDs[chat.ID] = true
			}
		}

		hub.register <- client

		go client.writePump()
		go client.readPump(hub, messageStore, chatStore)
	}
}

func (c *Client) readPump(hub *Hub, messageStore types.MessageStore, chatStore types.ChatStore) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var payload types.MessagePayload
		err := c.conn.ReadJSON(&payload)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Websocket read error: %v", err)
			}
			break
		}

		if c.role == "teacher" {
			message := &types.Message{
				ChatID:    payload.Message.ChatID,
				SenderID:  c.userID,
				Content:   payload.Message.Content,
				CreatedAt: time.Now(),
			}

			id, err := messageStore.SaveMessage(message)
			if err != nil {
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to save message"))
				return
			}

			message.ID = id
			message.TeacherID = c.userID

			hub.broadcast <- *message
		} else if c.role == "student" {
			chat, err := chatStore.GetChatByUserIDs(c.userID, payload.UserID)
			if err != nil {
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to retrieve chat"))
				return
			}
			if chat == nil {
				chat = &types.Chat{
					ChatType: "",
					Name:     "",
				}
				err = chatStore.CreateChat(chat, []int{c.userID, payload.UserID})
				if err != nil {
					log.Println("Unable to create new chat:", err)
					c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to create chat"))
					return
				}
			}
			payload.Message.ChatID = chat.ID
			payload.Message.SenderID = c.userID

			id, err := messageStore.SaveMessage(&payload.Message)
			if err != nil {
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to save message"))
				return
			}

			payload.Message.ID = id
			fmt.Println(payload.Message)
			hub.broadcast <- payload.Message
		}
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
