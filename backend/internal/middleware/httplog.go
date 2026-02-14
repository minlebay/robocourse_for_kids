package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPLog returns a Gin middleware that logs each request with method, path, status, duration and request_id (slog).
func HTTPLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}
		clientIP := c.ClientIP()
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)
		rid, _ := c.Get("request_id")

		slog.Info("request",
			"method", method,
			"path", path,
			"status", status,
			"ip", clientIP,
			"latency_ms", latency.Milliseconds(),
			"request_id", rid,
		)
	}
}
