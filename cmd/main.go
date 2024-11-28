package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/cache"
	"github.com/prok05/ecom/cmd/api"
	"github.com/prok05/ecom/db"
	"github.com/prok05/ecom/service/ws"
	"log"
	"time"
)

func main() {
	dbpool, err := db.NewPostgreSQLStorage()
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()
	initStorage(dbpool)

	hub := ws.NewHub()
	go hub.Run()

	tokenCache := cache.NewTokenCache(60 * time.Minute)
	_, err = tokenCache.GetToken()
	if err != nil {
		log.Fatal("cant initialize token cache")
	}
	log.Println("Token cache initialized")

	server := api.NewAPIServer(":8080", dbpool, hub, tokenCache)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(dbpool *pgxpool.Pool) {
	err := dbpool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DB: Successfully connected")
}
