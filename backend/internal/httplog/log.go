package httplog

import (
	"log"

	"github.com/gin-gonic/gin"
)

// LogError logs err with request id from context (if set). Use when returning 5xx to client.
func LogError(c *gin.Context, err error) {
	rid, _ := c.Get("request_id")
	log.Printf("[%v] %v", rid, err)
}
