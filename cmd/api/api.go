package api

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/cache"
	"github.com/prok05/ecom/service/chat"
	"github.com/prok05/ecom/service/homework"
	"github.com/prok05/ecom/service/lesson"
	"github.com/prok05/ecom/service/message"
	"github.com/prok05/ecom/service/user"
	"github.com/prok05/ecom/service/ws"
	"github.com/rs/cors"
	"log"
	"net/http"
)

type APIServer struct {
	addr       string
	dbpool     *pgxpool.Pool
	hub        *ws.Hub
	tokenCache *cache.TokenCache
}

func NewAPIServer(addr string, dbpool *pgxpool.Pool, hub *ws.Hub, tokenCache *cache.TokenCache) *APIServer {
	return &APIServer{
		addr:       addr,
		dbpool:     dbpool,
		hub:        hub,
		tokenCache: tokenCache,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.dbpool)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)

	messageStore := message.NewStore(s.dbpool)
	messageHandler := message.NewHandler(messageStore, s.tokenCache)
	messageHandler.RegisterRoutes(subrouter)

	chatStore := chat.NewStore(s.dbpool)
	chatHandler := chat.NewHandler(chatStore, s.tokenCache)
	chatHandler.RegisterRoutes(subrouter)

	homeworkStore := homework.NewStore(s.dbpool)
	homeworkHandler := homework.NewHandler(homeworkStore)
	homeworkHandler.RegisterRoutes(subrouter)

	lessonHandler := lesson.NewHandler(homeworkStore)
	lessonHandler.RegisterRoutes(subrouter)

	router.HandleFunc("/ws", ws.Handler(s.hub, messageStore))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"}, // Разрешаем запросы только с фронтенда
		//AllowedOrigins:   []string{"http://93.183.81.6:3000"},                          // Разрешаем запросы только с фронтенда
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}, // Разрешаемые методы
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Content-Disposition"}, // Разрешаемые заголовки
		AllowCredentials: true,                            // Разрешение для отправки куков и других креденшалов
	})

	handler := c.Handler(router)

	log.Println("Listening on:", s.addr)

	return http.ListenAndServe(s.addr, handler)
}
