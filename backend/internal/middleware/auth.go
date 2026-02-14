package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/domain/users"
)

// Auth parses JWT from the Authorization header and sets user_id and user_role
// in the Gin context. If no header is present, the request continues as anonymous.
// If the header is present but the token is invalid, returns 401.
func Auth(userHandler *users.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.Next()
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(auth, prefix)
		userID, role, err := userHandler.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Set("user_role", role)
		c.Next()
	}
}

// RequireAuth aborts with 401 if the user is not authenticated.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireTeacher aborts with 403 if the authenticated user is not a teacher.
// Must be used after RequireAuth.
func RequireTeacher() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		if role != users.RoleTeacher {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: teacher role required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserID returns the current user's ID from context, or uuid.Nil if not set.
func UserID(c *gin.Context) uuid.UUID {
	uid, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil
	}
	id, ok := uid.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}
