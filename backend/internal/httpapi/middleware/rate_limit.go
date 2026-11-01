package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

type IPRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	limit   int
	window  time.Duration
}

func NewIPRateLimiter(
	limit int,
	window time.Duration,
) *IPRateLimiter {
	return &IPRateLimiter{
		entries: make(map[string]*rateLimitEntry),
		limit:   limit,
		window:  window,
	}
}

func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if l.allow(c.ClientIP()) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(
			http.StatusTooManyRequests,
			gin.H{
				"error": "too many requests",
			},
		)
	}
}

func (l *IPRateLimiter) allow(ip string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanupExpired(now)

	entry, exists := l.entries[ip]

	if !exists {
		l.entries[ip] = &rateLimitEntry{
			count:   1,
			resetAt: now.Add(l.window),
		}

		return true
	}

	if !now.Before(entry.resetAt) {
		entry.count = 1
		entry.resetAt = now.Add(l.window)

		return true
	}

	if entry.count >= l.limit {
		return false
	}

	entry.count++

	return true
}

func (l *IPRateLimiter) cleanupExpired(now time.Time) {
	for ip, entry := range l.entries {
		if !now.Before(entry.resetAt) {
			delete(l.entries, ip)
		}
	}
}
