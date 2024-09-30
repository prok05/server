package api

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/service/lesson"
	"github.com/prok05/ecom/service/message"
	"github.com/prok05/ecom/service/user"
	"github.com/prok05/ecom/service/ws"
	"github.com/rs/cors"
	"log"
	"net/http"
)

type APIServer struct {
	addr   string
	dbpool *pgxpool.Pool
	hub    *ws.Hub
}

func NewAPIServer(addr string, dbpool *pgxpool.Pool, hub *ws.Hub) *APIServer {
	return &APIServer{
		addr:   addr,
		dbpool: dbpool,
		hub:    hub,
	}
}

func (s *APIServer) Run() error {
	// здесь можно будет поменять роутер
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.dbpool)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)
	lesson.RegisterRoutes(subrouter)

	messageStore := message.NewStore(s.dbpool)
	messageHandler := message.NewHandler(messageStore)
	messageHandler.RegisterRoutes(subrouter)

	router.HandleFunc("/ws", ws.Handler(s.hub, messageStore))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},                   // Разрешаем запросы только с фронтенда
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Разрешаемые методы
		AllowedHeaders:   []string{"Authorization", "Content-Type"},           // Разрешаемые заголовки
		AllowCredentials: true,                                                // Разрешение для отправки куков и других креденшалов
	})

	handler := c.Handler(router)

	log.Println(" Listening on:", s.addr)

	return http.ListenAndServe(s.addr, handler)
}
