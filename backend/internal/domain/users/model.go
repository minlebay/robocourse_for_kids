package users

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleStudent = "student"
	RoleTeacher = "teacher"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type UserWithPassword struct {
	User
	PasswordHash string
}
