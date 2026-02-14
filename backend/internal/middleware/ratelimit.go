package middleware

import (
	"context"
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
// If shutdown is not nil, the cleanup goroutine stops when shutdown is cancelled (e.g. on server shutdown).
func NewRateLimiter(limit int, window time.Duration, shutdown context.Context) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*ipEntry),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup(shutdown)
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
// If shutdown != nil, exits when shutdown is done (graceful server stop).
func (rl *RateLimiter) cleanup(shutdown context.Context) {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	doCleanup := func() {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.entries {
			if now.After(entry.resetAt) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
	if shutdown == nil {
		for range ticker.C {
			doCleanup()
		}
		return
	}
	for {
		select {
		case <-ticker.C:
			doCleanup()
		case <-shutdown.Done():
			return
		}
	}
}
