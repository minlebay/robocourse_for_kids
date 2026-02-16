package requestcontext

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const UserIDKey = "user_id"

// GetUserID returns the current user's ID from the Gin context, or uuid.Nil if not set or invalid.
// Used by handlers to avoid importing middleware (breaks import cycle with users).
func GetUserID(c *gin.Context) uuid.UUID {
	uid, ok := c.Get(UserIDKey)
	if !ok {
		return uuid.Nil
	}
	id, ok := uid.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}
