package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/domain/users"
)

func Auth(userHandler *users.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.Next()
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			c.Next()
			return
		}
		token := strings.TrimPrefix(auth, prefix)
		userID, _, err := userHandler.ParseToken(token)
		if err != nil {
			c.Next()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

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

// UserID returns the current user's ID from context, or uuid.Nil if not set.
func UserID(c *gin.Context) uuid.UUID {
	uid, _ := c.Get("user_id")
	if uid == nil {
		return uuid.Nil
	}
	return uid.(uuid.UUID)
}
