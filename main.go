package main

import (
	"fmt"
	"log"
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
	store, err := NewStore()

	if err != nil {
		log.Fatal(err)
	}

	initErr := InitStore(store)

	if initErr != nil {
		log.Fatal(initErr)
	}
	fmt.Printf("%v", store.db)
}
