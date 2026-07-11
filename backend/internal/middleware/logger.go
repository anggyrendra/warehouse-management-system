package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs each incoming request with method, path, status, latency,
// and client IP. This is the central request-logging middleware.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[REQUEST] %d | %13v | %15s | %-7s %s",
			status, latency, clientIP, c.Request.Method, path)
	}
}
