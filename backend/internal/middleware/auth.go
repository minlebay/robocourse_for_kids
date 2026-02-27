package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/domain/users"
	"learn_kids/backend/internal/requestcontext"
)

const newTokenHeader = "X-New-Token"

// AuthProvider is used by Auth middleware to parse tokens and check user status.
// *users.Handler implements this interface.
type AuthProvider interface {
	ParseToken(tokenString string) (userID uuid.UUID, roles []string, mustChangePassword bool, expiresAt time.Time, err error)
	IsUserBlocked(ctx context.Context, id uuid.UUID) (bool, error)
	NewToken(userID uuid.UUID, roles []string, mustChangePassword bool) (string, error)
}

// Auth parses JWT from the Authorization header and sets user_id and user_role
// in the Gin context. If no header is present, the request continues as anonymous.
// If the header is present but the token is invalid, returns 401.
// Sliding session: если до истечения токена осталось меньше SlidingRefreshThreshold,
// в ответ добавляется заголовок X-New-Token с новым токеном.
func Auth(provider AuthProvider) gin.HandlerFunc {
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
		userID, roles, mustChangePassword, expiresAt, err := provider.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}
		// Проверяем наличие пользователя и is_blocked при каждом запросе:
		// удалённый пользователь или блокировка сразу дают отказ.
		blocked, err := provider.IsUserBlocked(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}
		if blocked {
			c.JSON(http.StatusForbidden, gin.H{"error": "user is blocked"})
			c.Abort()
			return
		}
		c.Set(requestcontext.UserIDKey, userID)
		c.Set("user_roles", roles)
		c.Set("must_change_password", mustChangePassword)
		// Sliding session: при остатке времени меньше порога выдаём новый токен
		if !expiresAt.IsZero() && time.Until(expiresAt) < users.SlidingRefreshThreshold {
			if newToken, err := provider.NewToken(userID, roles, mustChangePassword); err == nil {
				c.Header(newTokenHeader, newToken)
			}
		}
		c.Next()
	}
}

// RequireAuth aborts with 401 if the user is not authenticated.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(requestcontext.UserIDKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func hasRole(roles []string, want string) bool {
	for _, r := range roles {
		if r == want {
			return true
		}
	}
	return false
}

// RequireTeacher aborts with 403 if the authenticated user is not a teacher or administrator.
// Must be used after RequireAuth.
func RequireTeacher() gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesVal, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		roles, _ := rolesVal.([]string)
		if !hasRole(roles, users.RoleTeacher) && !hasRole(roles, users.RoleAdministrator) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdmin aborts with 403 if the authenticated user is not an administrator.
// Must be used after RequireAuth.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesVal, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		roles, _ := rolesVal.([]string)
		if !hasRole(roles, users.RoleAdministrator) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireFreshPassword blocks all requests if must_change_password is true,
// except for POST /api/v1/auth/change-password.
func RequireFreshPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		mustChange, exists := c.Get("must_change_password")
		if exists && mustChange == true {
			if c.FullPath() != "/api/v1/auth/change-password" {
				c.JSON(http.StatusForbidden, gin.H{"error": "password_change_required"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// UserID returns the current user's ID from context, or uuid.Nil if not set.
func UserID(c *gin.Context) uuid.UUID {
	return requestcontext.GetUserID(c)
}
