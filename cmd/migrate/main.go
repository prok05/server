package main

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/prok05/ecom/config"
	"github.com/prok05/ecom/db"
	"log"
	"os"
)

func main() {
	dbpool, err := db.NewPostgreSQLStorage()
	if err != nil {
		log.Fatal(err)
	}

	sqlDB := stdlib.OpenDBFromPool(dbpool)

	driver, err := pgx.WithInstance(sqlDB, &pgx.Config{})
	if err != nil {
		log.Fatalf("Could not create migration driver: %v\n", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
		config.Envs.DBName,
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	cmd := os.Args[len(os.Args)-1]
	if cmd == "up" {
		log.Println("Starting migration UP...")
		err := m.Up()
		if err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("No change in migrations, everything is up-to-date")
			} else {
				log.Fatalf("Migration UP failed: %v", err)
			}
		} else {
			log.Println("Migration UP completed successfully")
		}
	}
	if cmd == "down" {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
	}
}
