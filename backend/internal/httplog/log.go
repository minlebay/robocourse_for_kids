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
