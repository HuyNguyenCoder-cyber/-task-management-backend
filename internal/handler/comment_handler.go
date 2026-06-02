package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"task-management-backend/internal/auth"
	"task-management-backend/internal/domain"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/service"
	"task-management-backend/internal/websocket"
	"task-management-backend/internal/worker"
)

type CommentHandler struct {
	taskService    service.TaskService
	commentRepo    repository.CommentRepository
	userRepository repository.UserRepository
	redisClient    *redis.Client
}

func NewCommentHandler(
	taskService service.TaskService,
	commentRepo repository.CommentRepository,
	userRepository repository.UserRepository,
	redisClient *redis.Client,
) *CommentHandler {
	return &CommentHandler{
		taskService:    taskService,
		commentRepo:    commentRepo,
		userRepository: userRepository,
		redisClient:    redisClient,
	}
}

type createCommentRequest struct {
	Content string `json:"content"`
}

type commentResponse struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *CommentHandler) ListTaskComments(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleServiceError(c, service.ErrUnauthorized)
		return
	}

	taskID, ok := parseInt64Param(c, "id", "Invalid task ID")
	if !ok {
		return
	}

	if _, err := h.taskService.GetTaskByID(c.Request.Context(), userID, taskID); err != nil {
		handleServiceError(c, err)
		return
	}

	comments, err := h.commentRepo.FindByTaskID(c.Request.Context(), taskID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list comments", err)
		return
	}

	responses := make([]commentResponse, 0, len(comments))
	for _, comment := range comments {
		responses = append(responses, commentResponse{
			ID:        comment.ID,
			TaskID:    comment.TaskID,
			UserID:    comment.UserID,
			Username:  h.resolveUsername(c, comment.UserID),
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		})
	}

	writeSuccess(c, http.StatusOK, "Comments retrieved successfully", responses)
}

func (h *CommentHandler) CreateTaskComment(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleServiceError(c, service.ErrUnauthorized)
		return
	}

	taskID, ok := parseInt64Param(c, "id", "Invalid task ID")
	if !ok {
		return
	}

	task, err := h.taskService.GetTaskByID(c.Request.Context(), userID, taskID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		writeError(c, http.StatusBadRequest, "Comment content is required", nil)
		return
	}

	comment := &domain.Comment{
		TaskID:  taskID,
		UserID:  userID,
		Content: strings.TrimSpace(req.Content),
	}
	if err := h.commentRepo.Create(c.Request.Context(), comment); err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to create comment", err)
		return
	}

	username := h.resolveUsername(c, userID)

	// Always enqueue one notification job for comment events (debug console flow like task assignment).
	notifyUserID := task.CreatedBy
	if task.AssigneeID != nil {
		notifyUserID = *task.AssigneeID
	}
	if notifyUserID == 0 {
		notifyUserID = userID
	}
	worker.EnqueueNotificationJob(c.Request.Context(), h.redisClient, taskID, notifyUserID)

	websocket.PublishEvent(websocket.CommentEvent{
		Event:     "comment_created",
		TaskID:    uint(comment.TaskID),
		CommentID: uint(comment.ID),
		Content:   comment.Content,
		UserID:    uint(comment.UserID),
		Username:  username,
	})

	writeSuccess(c, http.StatusCreated, "Comment created successfully", commentResponse{
		ID:        comment.ID,
		TaskID:    comment.TaskID,
		UserID:    comment.UserID,
		Username:  username,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	})
}

func (h *CommentHandler) resolveUsername(c *gin.Context, userID int64) string {
	username := "Unknown user"
	if currentUser, err := h.userRepository.GetByID(c.Request.Context(), userID); err == nil {
		username = strings.TrimSpace(currentUser.FullName)
		if username == "" {
			username = strings.TrimSpace(currentUser.Email)
		}
	}
	return username
}
