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

	teacher := NewTeacher("teacher@example.com")
	mock.ExpectExec("INSERT INTO teachers").WithArgs(teacher.Email).WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.AddTeacher(teacher)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddStudents(t *testing.T) {
	db, mock := NewMockDB()
	defer db.Close()

	store := &Store{db: db}

	students := []*Student{NewStudent("student1@gmail.com"), NewStudent("student2@gmail.com")}

	mock.ExpectExec("INSERT INTO students \\(email\\)").WithArgs(students[0].Email, students[1].Email).WillReturnResult(sqlmock.NewResult(1, int64(len(students))))

	err := store.AddStudents(students)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister(t *testing.T) {
	db, mock := NewMockDB()
	defer db.Close()

	store := &Store{db: db}
	pairs := []*TeacherStudentPair{NewTeacherStudentPair("teacher@gmail.com", "student1@gmail.com"), NewTeacherStudentPair("teacher@gmail.com", "student2@gmail.com")}

	mock.ExpectExec("INSERT INTO registered \\(student_email, teacher_email\\) VALUES").
		WithArgs(pairs[0].StudentEmail, pairs[0].TeacherEmail, pairs[1].StudentEmail, pairs[1].TeacherEmail).
		WillReturnResult(sqlmock.NewResult(1, 2))

	err := store.Register(pairs)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddSuspension(t *testing.T) {
	db, mock := NewMockDB()
	defer db.Close()

	store := &Store{db: db}

	suspension := NewSuspension("student_suspened@example.com")
	mock.ExpectExec("INSERT INTO suspensions").WithArgs(suspension.Email, suspension.SuspendedAt).WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.AddSuspension(suspension)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
