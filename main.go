package main

import (
	"log"

	"net/http"

	"github.com/gin-gonic/gin"
)

type Teacher struct {
	Email    string `json:"email"`
	Students []Student
}

type Student struct {
	Email    string `json:"email"`
	Teachers []Teacher
}

type TeacherStudentPair struct {
	TeacherEmail string `json:"teacherEmail"`
	StudentEmail string `json:"studentEmail"`
}

type apiHandler func(c *gin.Context, store *Store)

func main() {
	// Instantiate and init store

	store, err := NewStore()

	if err != nil {
		log.Fatal(err)
	}

	initErr := store.Init()

	if initErr != nil {
		log.Fatal(initErr)
	}

	// Setup and run the server
	router := setupRouter(store)
	router.Run(":8080")
}

func setupRouter(store *Store) *gin.Engine {
	router := gin.Default()
	router.POST("/api/register", makeHandleFunc(handleRegister, store))

	return router
}

func handleRegister(c *gin.Context, store *Store) {
	var input struct {
		Teacher  string   `json:"teacher" binding:"required"`
		Students []string `json:"students"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add teacher if does not exist
	if err := store.AddTeacher(input.Teacher); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "failed to add teacher."})
		return
	}

	// Add students if do not exist
	if err := store.AddStudents(input.Students); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "failed to add students."})
		return
	}

	// Register students to Teacher
	teacherStudentPairs := []TeacherStudentPair{}
	for _, student := range input.Students {
		teacherStudentPairs = append(teacherStudentPairs, TeacherStudentPair{input.Teacher, student})
	}
	if err := store.Register(teacherStudentPairs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "failed to register students to teachers."})
		return
	}

	c.Status(http.StatusNoContent)
}

// Function to convert API Handlers to Gin Handle Funcs because of the store param
func makeHandleFunc(apiHandler apiHandler, store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiHandler(c, store)
	}
}
