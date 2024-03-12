package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

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
	router.POST("/api/retrievefornotifications", makeHandleFunc(handleRetrieveNotifications, store))

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

	// Check if student is registered
	isStudentExists, err := store.IfStudentExists(input.Student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Something went wrong when checking if student is registered."})
		return
	}
	if !isStudentExists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Given student is not registered."})
		return
	}

	// validate email and create new Suspension instance
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

func handleRetrieveNotifications(c *gin.Context, store *Store) {
	var input struct {
		Teacher      string `json:"teacher" binding:"required"`
		Notification string `json:"notification" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "One or more fields are missing or invalid."})
		return
	}

	// Check if teacher is registered
	isTeacherExists, err := store.IfTeacherExists(input.Teacher)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Something went wrong when checking if teacher is registered."})
		return
	}
	if !isTeacherExists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Given teacher is not registered."})
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

	// Handling of mentioned emails
	// Get all mentioned emails
	emailPattern := `@\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`

	re := regexp.MustCompile(emailPattern)
	emails := re.FindAllString(input.Notification, -1)

	notifiableEmailsMap := map[string]bool{}

	for _, email := range emails {
		// Strip the @ of the mention
		email = email[1:]
		// Check if mentioned email is a student
		isStudentExists, err := store.IfStudentExists(email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": fmt.Sprintf("Something went wrong when checking if %s is registered.", email)})
			return
		}
		if !isStudentExists {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Student (%s) mentioned is not registered.", email)})
			return
		}

		// Check if mentioned email is suspended
		isSuspended, err := store.IsSuspended(email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": fmt.Sprintf("Something went wrong when checking if %s is suspended.", email)})
			return
		}
		if !isSuspended {
			notifiableEmailsMap[email] = true
		}
	}

	// Handling of students registered to teacher
	studentEmails, err := store.GetNotifiableStudentsOfTeacher(teacher)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Something went wrong when getting notifiable students"})
		return
	}

	for _, studentEmail := range studentEmails {
		notifiableEmailsMap[studentEmail] = true
	}

	// Get final slice of notifiable emails
	notifiableEmails := []string{}
	for notifiableEmail := range notifiableEmailsMap {
		notifiableEmails = append(notifiableEmails, notifiableEmail)
	}

	c.JSON(http.StatusOK, gin.H{"recipients": notifiableEmails})
}

// Function to convert API Handlers to Gin Handle Funcs because of the store param
func makeHandleFunc(apiHandler apiHandler, store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiHandler(c, store)
	}
}
