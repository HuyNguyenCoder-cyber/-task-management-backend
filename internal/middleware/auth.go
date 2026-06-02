package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/auth"
)

func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
				"error":   "authorization header is required",
			})
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
				"error":   "authorization header must be Bearer token",
			})
			return
		}

		claims, err := jwtService.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized",
				"error":   "invalid or expired token",
			})
			return
		}

		ctx := auth.ContextWithUserID(c.Request.Context(), claims.UserID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
