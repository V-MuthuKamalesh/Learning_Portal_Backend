package middleware

import (
	"sync"
	"time"

	"github.com/collegeassess/backend/pkg/logger"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware assigns a correlation id and echoes it in the response header.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(ctxRequestID, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// Logger logs each request with method, path, status and latency.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("request",
			"id", RequestID(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}

// Recovery converts panics into a 500 JSON error.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered", "error", r, "path", c.Request.URL.Path)
				response.Internal(c, "internal server error")
			}
		}()
		c.Next()
	}
}

// CORS configures allowed origins for the SPA portals.
func CORS(origins []string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// RateLimit is a simple in-memory sliding-window limiter keyed by client IP.
// For multi-instance deployments, back this with Redis.
func RateLimit(maxPerMinute int) gin.HandlerFunc {
	type bucket struct {
		count int
		reset time.Time
	}
	var mu sync.Mutex
	buckets := map[string]*bucket{}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		b, ok := buckets[ip]
		if !ok || now.After(b.reset) {
			b = &bucket{count: 0, reset: now.Add(time.Minute)}
			buckets[ip] = b
		}
		b.count++
		over := b.count > maxPerMinute
		mu.Unlock()

		if over {
			response.Error(c, 429, "rate_limited", "too many requests, slow down")
			return
		}
		c.Next()
	}
}
