package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Store struct {
	db *sqlx.DB
}

func NewStore() (*Store, error) {
	// Load .env file for POSTGRES conn details.
	LoadDotEnv()

	dbUser := GetEnv("POSTGRES_USER")
	dbName := GetEnv("POSTGRES_DB")
	dbPassword := GetEnv("POSTGRES_PASSWORD")

	connStr := fmt.Sprintf("user=%v dbname=%v sslmode=disable password=%v host=localhost", dbUser, dbName, dbPassword)
	db, err := sqlx.Connect("postgres", connStr)

	if err != nil {
		return nil, err
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Printf("Connected to Postgres database: %v", dbName)
	return &Store{db: db}, nil
}
