package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadDotEnv() {
	err := godotenv.Load(".env")

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
