package main

import (
	"fmt"
	"log"
)

type Teacher struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Students []Student
}

type Student struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Teachers []Teacher
}

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
