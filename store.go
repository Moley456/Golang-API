package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

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
	dbPort := GetEnv("POSTGRES_PORT")
	dbHost := GetEnv("POSTGRES_HOST")

	connStr := fmt.Sprintf("user=%v dbname=%v sslmode=disable password=%v host=%v port=%v", dbUser, dbName, dbPassword, dbHost, dbPort)
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

func (store *Store) Init() error {
	err1 := store.createStudentTable()
	err2 := store.createTeacherTable()
	err3 := store.createRegisteredTable()
	err4 := store.createSuspendedTable()
	err := errors.Join(err1, err2, err3, err4)
	return err
}

func (store *Store) createStudentTable() error {
	query := `CREATE TABLE IF NOT EXISTS students(
		email VARCHAR(50) PRIMARY KEY
	)`

	_, err := store.db.Exec(query)
	return err
}

func (store *Store) createTeacherTable() error {
	query := `CREATE TABLE IF NOT EXISTS teachers(
		email VARCHAR(50) PRIMARY KEY
	)`

	_, err := store.db.Exec(query)
	return err
}

func (store *Store) createRegisteredTable() error {
	query := `CREATE TABLE IF NOT EXISTS registered(
		student_email  VARCHAR(50),
		teacher_email  VARCHAR(50),
		PRIMARY KEY (student_email, teacher_email),
		FOREIGN KEY (student_email) REFERENCES students(email) ON DELETE CASCADE,
		FOREIGN KEY (teacher_email) REFERENCES teachers(email) ON DELETE CASCADE
	)`

	_, err := store.db.Exec(query)
	return err
}

func (store *Store) createSuspendedTable() error {
	query := `CREATE TABLE IF NOT EXISTS registered(
		student_email  VARCHAR(50),
		suspended_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (student_email, suspended_at),
		FOREIGN KEY (student_email) REFERENCES students(email) ON DELETE CASCADE
	)`

	_, err := store.db.Exec(query)
	return err
}

func (store *Store) AddTeacher(teacher *Teacher) error {
	query := `INSERT INTO teachers (email) VALUES ($1) ON CONFLICT DO NOTHING`

	_, err := store.db.Exec(query, teacher.Email)
	return err
}

func (store *Store) AddStudents(students []*Student) error {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("INSERT INTO students (email) VALUES ")
	params := make([]interface{}, len(students))
	for i, student := range students {
		params[i] = student.Email

		queryBuilder.WriteString(fmt.Sprintf("($%d),", i+1))
	}

	// drop last comma
	query := queryBuilder.String()
	query = query[:len(query)-1] + " ON CONFLICT DO NOTHING"

	_, err := store.db.Exec(query, params...)
	return err
}

func (store *Store) Register(teacherStudentPairs []*TeacherStudentPair) error {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("INSERT INTO registered (student_email, teacher_email) VALUES ")
	params := make([]interface{}, len(teacherStudentPairs)*2)
	for i, pair := range teacherStudentPairs {
		pos := i * 2
		params[pos] = pair.StudentEmail
		params[pos+1] = pair.TeacherEmail

		queryBuilder.WriteString(fmt.Sprintf("($%d, $%d),", pos+1, pos+2))
	}

	// drop last comma
	query := queryBuilder.String()
	query = query[:len(query)-1] + " ON CONFLICT DO NOTHING"

	_, err := store.db.Exec(query, params...)
	return err
}

func (store *Store) GetCommonStudents(teachers []string) ([]string, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`SELECT student_email AS email
		FROM registered
		WHERE teacher_email IN (`)
	params := make([]interface{}, len(teachers))
	for i, teacher := range teachers {
		params[i] = teacher

		queryBuilder.WriteString(fmt.Sprintf("$%d,", i+1))
	}

	query := queryBuilder.String()

	// remove the last comma and finish the query
	query = query[:len(query)-1] + fmt.Sprintf(") GROUP BY student_email HAVING COUNT(DISTINCT teacher_email) = $%d;", len(params)+1)

	// Append param for the HAVING condition
	params = append(params, len(teachers))
	students := []string{}

	err := store.db.Select(&students, query, params...)
	if err != nil {
		return nil, err
	}
	return students, nil
}
