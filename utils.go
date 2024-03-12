package main

import (
	"fmt"
	"log"
	"net/mail"
	"os"

	"github.com/joho/godotenv"
)

func LoadDotEnv() {
	// Load only .dev.env
	err := godotenv.Load(".dev.env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func GetEnv(key string) string {
	value, exists := os.LookupEnv(key)

	if !exists {
		log.Fatalf(fmt.Sprintf("Missing env var with key: %v", key))
	}

	return value
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
