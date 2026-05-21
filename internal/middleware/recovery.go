package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, exists := c.Get(RequestIDKey)
				if !exists {
					requestID = "-"
				}

				log.Printf("[PANIC] request_id=%v error=%v", requestID, err)

				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Internal server error",
					"error":   "panic recovered",
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}