package main

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func NewMockDB() (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return sqlxDB, mock
}

func TestAddTeacher(t *testing.T) {
	db, mock := NewMockDB()
	defer db.Close()

	store := &Store{db: db}

	teacherEmail := "teacher@example.com"
	mock.ExpectExec("INSERT INTO teachers").WithArgs(teacherEmail).WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.AddTeacher(teacherEmail)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddStudents(t *testing.T) {
	db, mock := NewMockDB()
	defer db.Close()

	store := &Store{db: db}

	studentEmails := []string{"student1@gmail.com", "student2@gmail.com"}

	values := make([]string, len(studentEmails))
	for i, email := range studentEmails {
		values[i] = fmt.Sprintf("\\('%s'\\)", email)
	}

	query := fmt.Sprintf("INSERT INTO students \\(email\\) VALUES %s ON CONFLICT DO NOTHING", strings.Join(values, ","))
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, int64(len(studentEmails))))

	err := store.AddStudents(studentEmails)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister(t *testing.T) {
	db, mock := NewMockDB()
	defer db.Close()

	store := &Store{db: db}
	pairs := []TeacherStudentPair{TeacherStudentPair{"teacher@gmail.com", "student1@gmail.com"}, TeacherStudentPair{"teacher@gmail.com", "student1@gmail.com"}}

	values := make([]string, len(pairs))
	for i, pair := range pairs {
		values[i] = fmt.Sprintf("\\('%s', '%s'\\)", pair.StudentEmail, pair.TeacherEmail)
	}

	query := fmt.Sprintf("INSERT INTO registered \\(student_email, teacher_email\\) VALUES %s ON CONFLICT DO NOTHING", strings.Join(values, ","))
	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 2))

	err := store.Register(pairs)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
