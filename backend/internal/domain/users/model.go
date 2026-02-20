package users

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleStudent       = "student"
	RoleTeacher       = "teacher"
	RoleAdministrator = "administrator"
)

type User struct {
	ID                 uuid.UUID `json:"id"`
	Login              string    `json:"login"`
	Name               string    `json:"name"`
	Role               string    `json:"role"`
	Theme              string    `json:"theme,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	Email              *string   `json:"email,omitempty"`
	MustChangePassword bool      `json:"must_change_password"`
	IsBlocked          bool      `json:"is_blocked"`
}

type UserWithPassword struct {
	User
	PasswordHash string
}
