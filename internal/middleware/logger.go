package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		requestID, exists := c.Get(RequestIDKey)
		if !exists {
			requestID = "-"
		}

		log.Printf(
			"[REQUEST] request_id=%v method=%s path=%s status=%d latency=%s client_ip=%s",
			requestID,
			method,
			path,
			statusCode,
			latency,
			clientIP,
		)
	}
}