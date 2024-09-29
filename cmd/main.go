package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/cmd/api"
	"github.com/prok05/ecom/db"
	"log"
)

func main() {
	// можно здесь инициализировать slog и потом передать его в структуру сервера api/

	dbpool, err := db.NewPostgreSQLStorage()
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()

	initStorage(dbpool)

	server := api.NewAPIServer(":8080", dbpool)
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
