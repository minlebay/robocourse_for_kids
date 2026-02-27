package users

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleStudent       = "student"
	RoleTeacher       = "teacher"
	RoleCourseOwner   = "course_owner"
	RoleAdministrator = "administrator"
)

type User struct {
	ID                 uuid.UUID `json:"id"`
	Login              string    `json:"login"`
	Name               string    `json:"name"`
	Role               string    `json:"role"` // primary role for API compat; derived from user_roles (source of truth)
	Roles              []string  `json:"roles,omitempty"` // from user_roles
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

// PrimaryRole returns a single "primary" role for API backward compatibility.
// Order: administrator > teacher > course_owner > student. Used when users.role is no longer written.
func PrimaryRole(roles []string) string {
	for _, r := range []string{RoleAdministrator, RoleTeacher, RoleCourseOwner, RoleStudent} {
		for _, have := range roles {
			if have == r {
				return r
			}
		}
	}
	return ""
}
