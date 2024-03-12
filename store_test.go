package main

import (
	"log"
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

	mock.ExpectExec("INSERT INTO students \\(email\\)").WithArgs(studentEmails[0], studentEmails[1]).WillReturnResult(sqlmock.NewResult(1, int64(len(studentEmails))))

	err := store.AddStudents(studentEmails)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister(t *testing.T) {
	db, mock := NewMockDB()
	defer db.Close()

	store := &Store{db: db}
	pairs := []TeacherStudentPair{{"teacher@gmail.com", "student1@gmail.com"}, {"teacher@gmail.com", "student2@gmail.com"}}

	mock.ExpectExec("INSERT INTO registered \\(student_email, teacher_email\\) VALUES").
		WithArgs("student1@gmail.com", "teacher@gmail.com", "student2@gmail.com", "teacher@gmail.com").
		WillReturnResult(sqlmock.NewResult(1, 2))

	err := store.Register(pairs)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
