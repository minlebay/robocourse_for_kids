package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipEntry struct {
	count   int
	resetAt time.Time
}

// RateLimiter is a simple fixed-window in-memory rate limiter per client IP.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
	limit   int
	window  time.Duration
}

// NewRateLimiter creates a rate limiter: at most limit requests per window per IP.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*ipEntry),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

// Handler returns a Gin middleware that enforces the rate limit.
func (rl *RateLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		rl.mu.Lock()
		entry, ok := rl.entries[ip]
		if !ok || now.After(entry.resetAt) {
			rl.entries[ip] = &ipEntry{count: 1, resetAt: now.Add(rl.window)}
			rl.mu.Unlock()
			c.Next()
			return
		}
		entry.count++
		if entry.count > rl.limit {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, try again later"})
			c.Abort()
			return
		}
		rl.mu.Unlock()
		c.Next()
	}
}

// cleanup periodically removes expired entries to prevent memory leaks.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.entries {
			if now.After(entry.resetAt) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}
