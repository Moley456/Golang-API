package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Helper function to create a new Store connected to your test DB
func newTestStore() *Store {
	// Connect to your test database here
	// Make sure to handle errors appropriately (not shown for brevity)
	store, err := NewStore()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	initErr := store.Init()

	if initErr != nil {
		log.Fatal(initErr)
	}

	return store
}

func cleanUp(store *Store) {
	store.db.Exec("DROP TABLE students CASCADE")
	store.db.Exec("DROP TABLE teachers CASCADE")
	store.db.Exec("DROP TABLE registered")
	store.db.Exec("DROP TABLE suspensions")
}

func TestHandleRegister(t *testing.T) {
	store := newTestStore()
	defer store.db.Close()
	router := SetupRouter(store)

	teacherEmail := "teacher@example.com"
	studentEmail1 := "student1@example.com"
	studentEmail2 := "student2@example.com"

	input := struct {
		Teacher  string   `json:"teacher"`
		Students []string `json:"students"`
	}{
		Teacher:  teacherEmail,
		Students: []string{studentEmail1, studentEmail2},
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	cleanUp(store)
}

func TestHandleCommonStudents(t *testing.T) {
	store := newTestStore()
	defer store.db.Close()
	router := SetupRouter(store)

	teacherEmail1 := "teacher1@example.com"
	teacherEmail2 := "teacher2@example.com"
	studentEmail1 := "student1@example.com"
	studentEmail2 := "student2@example.com"

	store.db.Exec("INSERT INTO teachers (email) VALUES ($1), ($2)", teacherEmail1, teacherEmail2)
	store.db.Exec("INSERT INTO students (email) VALUES ($1), ($2)", studentEmail1, studentEmail2)
	store.db.Exec("INSERT INTO registered (student_email, teacher_email) VALUES ($1, $2), ($3, $4), ($5, $6)",
		studentEmail1, teacherEmail1, studentEmail2, teacherEmail1, studentEmail1, teacherEmail2)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/commonstudents?teacher=%s&teacher=%s", teacherEmail1, teacherEmail2), nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected := struct {
		Students []string `json:"students"`
	}{
		Students: []string{studentEmail1},
	}

	expectedBytes, _ := json.Marshal(expected)
	require.Equal(t, expectedBytes, w.Body.Bytes())
	require.Equal(t, http.StatusOK, w.Code)
	cleanUp(store)
}

func TestHandleSuspension(t *testing.T) {
	store := newTestStore()
	defer store.db.Close()
	router := SetupRouter(store)

	studentEmail1 := "student1@example.com"
	store.db.Exec("INSERT INTO students (email) VALUES ($1)", studentEmail1)

	input := struct {
		Student string `json:"student"`
	}{
		Student: studentEmail1,
	}

	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/api/suspend", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	cleanUp(store)
}

func TestRetrieveNotificationNoSuspensionsWithDupes(t *testing.T) {
	store := newTestStore()
	defer store.db.Close()
	router := SetupRouter(store)

	teacherEmail1 := "teacher1@example.com"
	studentEmail1 := "student1@example.com"
	studentEmail2 := "student2@example.com"
	studentEmail3 := "student3@example.com"

	store.db.Exec("INSERT INTO teachers (email) VALUES ($1)", teacherEmail1)
	store.db.Exec("INSERT INTO students (email) VALUES ($1), ($2), ($3)", studentEmail1, studentEmail2, studentEmail3)
	store.db.Exec("INSERT INTO registered (student_email, teacher_email) VALUES ($1, $2), ($3, $4)",
		studentEmail1, teacherEmail1, studentEmail2, teacherEmail1)

	notification := fmt.Sprintf("Hello, @%s and @%s", studentEmail2, studentEmail3)
	input := struct {
		Teacher      string `json:"teacher"`
		Notification string `json:"notification"`
	}{
		Teacher:      teacherEmail1,
		Notification: notification,
	}
	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/api/retrievefornotifications", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected := struct {
		Recipients []string `json:"recipients"`
	}{
		Recipients: []string{studentEmail1, studentEmail2, studentEmail3},
	}

	sort.Strings(expected.Recipients)

	var res struct {
		Recipients []string `json:"recipients"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(res.Recipients)

	require.Equal(t, expected, res)
	require.Equal(t, http.StatusOK, w.Code)
	cleanUp(store)
}

func TestRetrieveNotificationWithSuspensions(t *testing.T) {
	store := newTestStore()
	defer store.db.Close()
	router := SetupRouter(store)

	teacherEmail1 := "teacher1@example.com"
	studentEmail1 := "student1@example.com"
	studentEmail2 := "student2@example.com"
	studentEmail3 := "student3@example.com"

	store.db.Exec("INSERT INTO teachers (email) VALUES ($1)", teacherEmail1)
	store.db.Exec("INSERT INTO students (email) VALUES ($1), ($2), ($3)", studentEmail1, studentEmail2, studentEmail3)
	store.db.Exec("INSERT INTO registered (student_email, teacher_email) VALUES ($1, $2), ($3, $4)",
		studentEmail1, teacherEmail1, studentEmail2, teacherEmail1)
	store.db.Exec("INSERT INTO suspensions (student_email, suspended_at) VALUES ($1, $2), ($3, $2)", studentEmail2, time.Now().UTC(), studentEmail3)

	notification := fmt.Sprintf("Hello, @%s", studentEmail3)
	input := struct {
		Teacher      string `json:"teacher"`
		Notification string `json:"notification"`
	}{
		Teacher:      teacherEmail1,
		Notification: notification,
	}
	data, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/api/retrievefornotifications", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected := struct {
		Recipients []string `json:"recipients"`
	}{
		Recipients: []string{studentEmail1},
	}

	expectedBytes, _ := json.Marshal(expected)
	require.Equal(t, expectedBytes, w.Body.Bytes())
	require.Equal(t, http.StatusOK, w.Code)
	cleanUp(store)
}
