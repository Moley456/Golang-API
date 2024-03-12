package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
	router.GET("/api/commonstudents", makeHandleFunc(handleCommonStudents, store))
	router.POST("/api/suspend", makeHandleFunc(handleSuspension, store))

	return router
}

func handleRegister(c *gin.Context, store *Store) {
	var input struct {
		Teacher  string   `json:"teacher" binding:"required"`
		Students []string `json:"students"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "One or more fields are missing or invalid."})
		return
	}

	// validate emails and create new teacher and student instances
	var teacher *Teacher
	if IsValidEmail(input.Teacher) {
		teacher = NewTeacher(input.Teacher)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Teacher's email (%s) is invalid.", input.Teacher)})
		return

	}

	students := []*Student{}
	for _, studentEmail := range input.Students {
		if IsValidEmail(studentEmail) {
			students = append(students, NewStudent(studentEmail))
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("A student's email (%s) is invalid.", studentEmail)})
			return
		}
	}

	// Add teacher if does not exist
	if err := store.AddTeacher(teacher); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "failed to add teacher."})
		return
	}

	if err := store.AddStudents(students); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "failed to add students."})
		return
	}

	// Register students to Teacher
	teacherStudentPairs := []*TeacherStudentPair{}
	for _, studentEmail := range input.Students {
		teacherStudentPairs = append(teacherStudentPairs, NewTeacherStudentPair(input.Teacher, studentEmail))
	}
	if err := store.Register(teacherStudentPairs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "failed to register students to teachers."})
		return
	}

	c.Status(http.StatusNoContent)
}

func handleCommonStudents(c *gin.Context, store *Store) {
	teacherEmails := c.QueryArray("teacher")
	if len(teacherEmails) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No teachers given in request."})
		return
	}

	// Validate emails and create teacher instances
	teachers := []*Teacher{}
	for _, teacherEmail := range teacherEmails {
		if IsValidEmail(teacherEmail) {
			teachers = append(teachers, NewTeacher(teacherEmail))
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("A teacher's email (%s) is invalid.", teacherEmail)})
			return
		}
	}

	students, err := store.GetCommonStudents(teachers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "failed to get common students."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"students": students})
}

func handleSuspension(c *gin.Context, store *Store) {
	var input struct {
		Student string `json:"student" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "student field is missing or invalid."})
		return
	}

	// validate email and create new Student instance
	var suspension *Suspension
	if IsValidEmail(input.Student) {
		suspension = NewSuspension(input.Student)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Student's email (%s) is invalid.", input.Student)})
		return
	}

	if err := store.AddSuspension(suspension); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "failed to suspend student."})
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
