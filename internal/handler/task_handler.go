package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/service"
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
	var req dto.CreateTaskRequest
//nếu bind thành công => err =nil => không bug
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	task, err := h.taskService.CreateTask(
		req.Title,
		req.Description,
		domain.TaskStatus(req.Status),
		req.Assignee,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusCreated, "Task created successfully", toTaskResponse(task))
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	tasks, err := h.taskService.ListTasks()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list tasks", err)
		return
	}

	responses := make([]dto.TaskResponse, 0, len(tasks))

	for _, task := range tasks {
		responses = append(responses, toTaskResponse(task))
	}

	writeSuccess(c, http.StatusOK, "Tasks retrieved successfully", responses)
}

func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	id := c.Param("id")

	task, err := h.taskService.GetTaskByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Task retrieved successfully", toTaskResponse(task))
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	task, err := h.taskService.UpdateTask(
		id,
		req.Title,
		req.Description,
		domain.TaskStatus(req.Status),
		req.Assignee,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Task updated successfully", toTaskResponse(task))
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	err := h.taskService.DeleteTask(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Task deleted successfully", nil)
}

func toTaskResponse(task *domain.Task) dto.TaskResponse {
	return dto.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Assignee:    task.Assignee,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrTaskTitleRequired):
		writeError(c, http.StatusBadRequest, "Task title is required", err)

	case errors.Is(err, service.ErrInvalidTaskStatus):
		writeError(c, http.StatusBadRequest, "Invalid task status", err)

	case errors.Is(err, repository.ErrTaskNotFound):
		writeError(c, http.StatusNotFound, "Task not found", err)

	default:
		writeError(c, http.StatusInternalServerError, "Internal server error", err)
	}
}