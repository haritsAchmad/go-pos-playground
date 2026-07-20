package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort          string
	JWTSecret        string
	JWTIssuer        string
	JWTExpiryMinutes string
	AdminName        string
	AdminEmail       string
	AdminPassword    string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSchema   string
	DBSSLMode  string
}

func Load() Config {

	err := godotenv.Load()

	if err != nil {
		log.Println(".env file not found, using system environment")
	}

	return Config{
		AppPort:          os.Getenv("APP_PORT"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		JWTIssuer:        os.Getenv("JWT_ISSUER"),
		JWTExpiryMinutes: os.Getenv("JWT_EXPIRY_MINUTES"),
		AdminName:        os.Getenv("INITIAL_ADMIN_NAME"),
		AdminEmail:       os.Getenv("INITIAL_ADMIN_EMAIL"),
		AdminPassword:    os.Getenv("INITIAL_ADMIN_PASSWORD"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSchema:   os.Getenv("DB_SCHEMA"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
	}
}
