package handler

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func writeSuccess(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func writeError(c *gin.Context, statusCode int, message string, err error) {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Error:   errorMessage,
	})
}