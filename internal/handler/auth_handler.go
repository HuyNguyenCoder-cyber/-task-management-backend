package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/dto"
	"task-management-backend/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	response, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusCreated, "User registered successfully", response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	response, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		handleAuthServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Login successful", response)
}

func handleAuthServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAuthEmailRequired),
		errors.Is(err, service.ErrAuthInvalidEmail),
		errors.Is(err, service.ErrAuthPasswordRequired),
		errors.Is(err, service.ErrAuthPasswordTooShort),
		errors.Is(err, service.ErrAuthFullNameRequired):
		writeError(c, http.StatusBadRequest, err.Error(), err)
	case errors.Is(err, service.ErrAuthEmailAlreadyExists):
		writeError(c, http.StatusConflict, "Email already exists", err)
	case errors.Is(err, service.ErrAuthInvalidCredentials):
		writeError(c, http.StatusUnauthorized, "Invalid email or password", err)
	default:
		writeError(c, http.StatusInternalServerError, "Internal server error", err)
	}
}
