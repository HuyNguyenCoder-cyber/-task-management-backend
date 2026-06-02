package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.ListUsers(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list users", err)
		return
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toUserResponse(user))
	}

	writeSuccess(c, http.StatusOK, "Users retrieved successfully", responses)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, ok := parseInt64Param(c, "id", "Invalid user ID")
	if !ok {
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "User retrieved successfully", toUserResponse(user))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := parseInt64Param(c, "id", "Invalid user ID")
	if !ok {
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), id, req.Email, req.FullName)
	if err != nil {
		handleUserServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "User updated successfully", toUserResponse(user))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, ok := parseInt64Param(c, "id", "Invalid user ID")
	if !ok {
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), id); err != nil {
		handleUserServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "User deleted successfully", nil)
}

func parseInt64Param(c *gin.Context, name string, message string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, http.StatusBadRequest, message, err)
		return 0, false
	}

	return id, true
}

func toUserResponse(user *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func handleUserServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserEmailRequired):
		writeError(c, http.StatusBadRequest, "User email is required", err)
	case errors.Is(err, service.ErrUserPasswordHashRequired):
		writeError(c, http.StatusBadRequest, "User password hash is required", err)
	case errors.Is(err, service.ErrUserFullNameRequired):
		writeError(c, http.StatusBadRequest, "User full name is required", err)
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(c, http.StatusNotFound, "User not found", err)
	default:
		writeError(c, http.StatusInternalServerError, "Internal server error", err)
	}
}
