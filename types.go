package main

import "time"

type Teacher struct {
	Email string `json:"email"`
}

func NewTeacher(email string) *Teacher {
	return &Teacher{
		Email: email,
	}
}

type Student struct {
	Email string `json:"email"`
}

func NewStudent(email string) *Student {
	return &Student{
		Email: email,
	}
}

type TeacherStudentPair struct {
	TeacherEmail string `json:"teacherEmail"`
	StudentEmail string `json:"studentEmail"`
}

func NewTeacherStudentPair(teacherEmail string, studentEmail string) *TeacherStudentPair {
	return &TeacherStudentPair{
		TeacherEmail: teacherEmail,
		StudentEmail: studentEmail,
	}
}

type Suspension struct {
	Email          string    `json:"email"`
	SuspendedAt    time.Time `json:"suspendedAt"`
	SuspendedUntil time.Time `json:"suspendedUntil"`
}

func NewSuspension(email string) *Suspension {
	return &Suspension{
		Email:       email,
		SuspendedAt: time.Now().UTC(),
	}
}
