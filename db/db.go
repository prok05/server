package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/config"
	"log"
)

func NewPostgreSQLStorage() (*pgxpool.Pool, error) {
	databaseUrl := fmt.Sprintf("%s://%s:%s@%s/%s",
		config.Envs.PublicHost,
		config.Envs.DBUser,
		config.Envs.DBPassword,
		config.Envs.DBAddress,
		config.Envs.DBName)

	dbConfig, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)

	if err != nil {
		log.Fatal(err)
	}

	return dbpool, nil
}
