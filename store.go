package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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
	err4 := store.createSuspensionTable()
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

func (store *Store) createSuspensionTable() error {
	query := `CREATE TABLE IF NOT EXISTS suspensions(
		student_email  VARCHAR(50),
		suspended_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		suspended_until TIMESTAMPTZ,
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

func (store *Store) IfStudentExists(email string) (bool, error) {
	var studentToFind []string
	query := `SELECT email FROM students WHERE email = $1`

	err := store.db.Select(&studentToFind, query, email)
	if err != nil {
		return false, err
	}

	return len(studentToFind) != 0, nil
}

func (store *Store) IfTeacherExists(email string) (bool, error) {
	var teacherToFind []string
	query := `SELECT email FROM teachers WHERE email = $1`

	err := store.db.Select(&teacherToFind, query, email)
	if err != nil {
		return false, err
	}

	return len(teacherToFind) != 0, nil
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

func (store *Store) GetCommonStudents(teachers []*Teacher) ([]string, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`SELECT student_email AS email FROM registered WHERE teacher_email IN (`)
	params := make([]interface{}, len(teachers))
	for i, teacher := range teachers {
		params[i] = teacher.Email

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

func (store *Store) AddSuspension(suspension *Suspension) error {
	query := `INSERT INTO suspensions (student_email, suspended_at) VALUES ($1, $2) ON CONFLICT DO NOTHING`

	_, err := store.db.Exec(query, suspension.Email, suspension.SuspendedAt)
	return err
}

func (store *Store) GetNotifiableStudentsOfTeacher(teacher *Teacher) ([]string, error) {
	query := `SELECT students.email FROM teachers 
	JOIN registered ON teachers.email=registered.teacher_email 
	JOIN students ON students.email=registered.student_email 
	JOIN suspensions ON students.email=suspensions.student_email
	WHERE teachers.email=$1 AND suspensions.suspended_at > $2 AND (suspensions.suspended_until < $2 OR suspensions.suspended_until IS NULL)`

	students := []string{}
	err := store.db.Select(&students, query, teacher.Email, time.Now().UTC())

	if err != nil {
		return nil, err
	}

	return students, nil
}

func (store *Store) IsSuspended(email string) (bool, error) {
	query := `SELECT student_email FROM suspensions
	WHERE student_email=$1 AND suspended_at <= $2 AND (suspended_until >= $2 OR suspended_until IS NULL)`

	students := []string{}
	err := store.db.Select(&students, query, email, time.Now().UTC())

	if err != nil {
		return false, err
	}

	return len(students) != 0, nil
}
