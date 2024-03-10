package main

import (
	"errors"
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

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Printf("Connected to Postgres database: %v", dbName)
	return &Store{db: db}, nil
}

func InitStore(store *Store) error {
	err1 := createStudentTable(store)
	err2 := createTeacherTable(store)
	err3 := createRegisteredTable(store)
	err := errors.Join(err1, err2, err3)
	return err
}

func createStudentTable(store *Store) error {
	query := `CREATE TABLE IF NOT EXISTS students(
		id INT PRIMARY KEY,
		email VARCHAR(50) UNIQUE NOT NULL
	)`

	_, err := store.db.Exec(query)
	return err
}

func createTeacherTable(store *Store) error {
	query := `CREATE TABLE IF NOT EXISTS teachers(
		id SERIAL PRIMARY KEY,
		email VARCHAR(50) UNIQUE NOT NULL
	)`

	_, err := store.db.Exec(query)
	return err
}

func createRegisteredTable(store *Store) error {
	query := `CREATE TABLE IF NOT EXISTS registered(
		student_id SERIAL,
		teacher_id SERIAL,
		PRIMARY KEY (student_id, teacher_id),
		FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
		FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE CASCADE
	)`

	_, err := store.db.Exec(query)
	return err
}
