package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prok05/ecom/db"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/service/user"
	"github.com/prok05/ecom/types"
	"log"
	"os"
)

func main() {
	phone := flag.String("phone", "", "Phone number for the supervisor")
	password := flag.String("password", "", "Password for the supervisor")

	flag.Parse()

	if *phone == "" || *password == "" {
		fmt.Println("Usage: go run main.go --phone=<phone> --password=<password>")
		os.Exit(1)
	}

	dbpool, err := db.NewPostgreSQLStorage()
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()
	initStorage(dbpool)

	hashedPassword, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("could not hash password: %v", err)
	}

	u := types.User{
		ID:         999999,
		FirstName:  "Админ",
		LastName:   "Админ",
		MiddleName: "Админ",
		Phone:      *phone,
		Password:   hashedPassword,
		Role:       "supervisor",
	}

	userStore := user.NewStore(dbpool)

	err = userStore.CreateUser(u)
	if err != nil {
		log.Fatalf("Error creating admin: %v", err)
	}

	log.Println("Admin created successfully!")
}

func initStorage(dbpool *pgxpool.Pool) {
	err := dbpool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DB: Successfully connected")
}
