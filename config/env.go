package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	PublicHost string
	Port       string

	DBUser     string
	DBPassword string
	DBAddress  string
	DBName     string

	AlphaEmail   string
	AlphaApiKey  string
	AlphaXAppKey string

	JWTExpirationInSeconds int64
	JWTSecret              string
}

var Envs = initConfig()

func initConfig() Config {
	path, err := os.Getwd()
	fmt.Println(filepath.Join(path, ".env"))
	if err != nil {
		log.Fatal("Error loading path")
	}

	err = godotenv.Load(filepath.Join(path, "../.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return Config{
		PublicHost: getEnv("PUBLIC_HOST", "postgres"),
		Port:       getEnv("PORT", "8080"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "1234"),
		DBAddress: fmt.Sprintf("%s:%s",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432")),
		DBName:                 getEnv("DB_NAME", "centriym-db"),
		AlphaEmail:             getEnv("ALPHA_EMAIL", "email"),
		AlphaApiKey:            getEnv("ALPHA_API_KEY", "api-key"),
		AlphaXAppKey:           getEnv("ALPHA_X_APP_KEY", "x-app-key"),
		JWTSecret:              getEnv("JWT_SECRET", "secret"),
		JWTExpirationInSeconds: getEnvAsInt("JWT_EXP", 3600*24*30),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
