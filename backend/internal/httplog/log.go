package httplog

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// LogError logs err with request id from context (if set). Use when returning 5xx to client.
func LogError(c *gin.Context, err error) {
	rid, _ := c.Get("request_id")
	slog.Error(err.Error(), "request_id", rid)
}

// LogWarn logs a warning with request id from context (if set). Use for non-critical issues.
func LogWarn(c *gin.Context, msg string) {
	rid, _ := c.Get("request_id")
	slog.Warn(msg, "request_id", rid)
}

// LogAudit logs a security-relevant action for audit trail purposes.
// Use for any destructive or privileged operation (delete, block, password reset, etc.).
func LogAudit(c *gin.Context, action string, args ...any) {
	rid, _ := c.Get("request_id")
	slog.Info("audit", append([]any{"action", action, "request_id", rid}, args...)...)
}
