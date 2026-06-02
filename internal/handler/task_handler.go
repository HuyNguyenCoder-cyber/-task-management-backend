package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/auth"
	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/service"
	"task-management-backend/internal/websocket"
)

type TaskHandler struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleServiceError(c, service.ErrUnauthorized)
		return
	}

	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	task, err := h.taskService.CreateTask(c.Request.Context(), userID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	websocket.PublishEvent(websocket.TaskEvent{
		Event:      "task_created",
		TaskID:     task.ID,
		ProjectID:  task.ProjectID,
		Title:      task.Title,
		Status:     string(task.Status),
		CreatedBy:  task.CreatedBy,
		AssigneeID: task.AssigneeID,
		DueDate:    task.DueDate,
	})

	writeSuccess(c, http.StatusCreated, "Task created successfully", toTaskResponse(task))
}

func (h *TaskHandler) CreateProjectTask(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleServiceError(c, service.ErrUnauthorized)
		return
	}

	projectID, ok := parseInt64Param(c, "id", "Invalid project ID")
	if !ok {
		return
	}

	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	task, err := h.taskService.CreateProjectTask(c.Request.Context(), userID, projectID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusCreated, "Task created successfully", toTaskResponse(task))
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleServiceError(c, service.ErrUnauthorized)
		return
	}

	projectID, ok := parseOptionalProjectID(c)
	if !ok {
		return
	}

	tasks, err := h.taskService.ListTasks(c.Request.Context(), userID, projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	responses := make([]dto.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, toTaskResponse(task))
	}

	writeSuccess(c, http.StatusOK, "Tasks retrieved successfully", responses)
}

func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleServiceError(c, service.ErrUnauthorized)
		return
	}

	id, ok := parseInt64Param(c, "id", "Invalid task ID")
	if !ok {
		return
	}

	task, err := h.taskService.GetTaskByID(c.Request.Context(), userID, id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Task retrieved successfully", toTaskResponse(task))
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleServiceError(c, service.ErrUnauthorized)
		return
	}

	id, ok := parseInt64Param(c, "id", "Invalid task ID")
	if !ok {
		return
	}

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	task, err := h.taskService.UpdateTask(c.Request.Context(), userID, id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	websocket.PublishEvent(websocket.TaskEvent{
		Event:  "task_updated",
		TaskID: task.ID,
		Status: string(task.Status),
	})

	writeSuccess(c, http.StatusOK, "Task updated successfully", toTaskResponse(task))
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleServiceError(c, service.ErrUnauthorized)
		return
	}

	id, ok := parseInt64Param(c, "id", "Invalid task ID")
	if !ok {
		return
	}

	err := h.taskService.DeleteTask(c.Request.Context(), userID, id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Task deleted successfully", nil)
}

func parseOptionalProjectID(c *gin.Context) (*int64, bool) {
	projectIDParam := c.Query("project_id")
	if projectIDParam == "" {
		return nil, true
	}

	projectID, err := strconv.ParseInt(projectIDParam, 10, 64)
	if err != nil || projectID <= 0 {
		writeError(c, http.StatusBadRequest, "Invalid project ID", err)
		return nil, false
	}

	return &projectID, true
}

func toTaskResponse(task *domain.Task) dto.TaskResponse {
	return dto.TaskResponse{
		ID:          task.ID,
		ProjectID:   task.ProjectID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		CreatedBy:   task.CreatedBy,
		AssigneeID:  task.AssigneeID,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUnauthorized):
		writeError(c, http.StatusUnauthorized, "Unauthorized", err)

	case errors.Is(err, service.ErrForbidden):
		writeError(c, http.StatusForbidden, "Forbidden", err)

	case errors.Is(err, service.ErrTaskTitleRequired):
		writeError(c, http.StatusBadRequest, "Task title is required", err)

	case errors.Is(err, service.ErrTaskProjectRequired):
		writeError(c, http.StatusBadRequest, "Task project is required", err)

	case errors.Is(err, service.ErrInvalidTaskStatus):
		writeError(c, http.StatusBadRequest, "Invalid task status", err)

	case errors.Is(err, repository.ErrProjectNotFound):
		writeError(c, http.StatusNotFound, "Project not found", err)

	case errors.Is(err, repository.ErrTaskNotFound):
		writeError(c, http.StatusNotFound, "Task not found", err)

	default:
		writeError(c, http.StatusInternalServerError, "Internal server error", err)
	}
}
