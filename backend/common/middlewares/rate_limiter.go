package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	sync.Mutex
	clients map[string][]time.Time
	limit   int
	window  time.Duration
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
}

// Middleware returns a Gin middleware that performs rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use user_id from context (set by Authenticate middleware) or fallback to IP
		userID, exists := c.Get("user_id")
		var key string
		if exists {
			key = userID.(string)
		} else {
			key = c.ClientIP()
		}

		rl.Lock()
		defer rl.Unlock()

		now := time.Now()
		cutoff := now.Add(-rl.window)

		// Clean up old requests for this client
		var validRequests []time.Time
		for _, t := range rl.clients[key] {
			if t.After(cutoff) {
				validRequests = append(validRequests, t)
			}
		}

		if len(validRequests) >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please try again later",
			})
			c.Abort()
			return
		}

		validRequests = append(validRequests, now)
		rl.clients[key] = validRequests

		c.Next()
	}
}
