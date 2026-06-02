package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"task-management-backend/internal/auth"
	"task-management-backend/internal/config"
	"task-management-backend/internal/database"
	"task-management-backend/internal/handler"
	"task-management-backend/internal/middleware"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/service"
)

type apiResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

type registerData struct {
	ID int64 `json:"id"`
}

type loginData struct {
	AccessToken string `json:"access_token"`
}

type projectData struct {
	ID int64 `json:"id"`
}

type taskData struct {
	ID int64 `json:"id"`
}

type commentData struct {
	ID      int64  `json:"id"`
	TaskID  int64  `json:"task_id"`
	Content string `json:"content"`
}

func performRequest(router *gin.Engine, method, path string, body any, token string) *httptest.ResponseRecorder {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func setupE2ERouter(t *testing.T) (*gin.Engine, *gorm.DB, *redis.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg, err := config.Load()
	if err != nil {
		t.Skipf("skip integration test: cannot load config: %v", err)
	}

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		t.Skipf("skip integration test: cannot connect database: %v", err)
	}

	redisClient, err := database.ConnectRedis(cfg.RedisAddr())
	if err != nil {
		t.Skipf("skip integration test: cannot connect redis: %v", err)
	}

	userRepo := repository.NewGormUserRepository(db)
	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiresInHours)
	authService := service.NewAuthService(userRepo, jwtService)
	authHandler := handler.NewAuthHandler(authService)

	projectMemberRepo := repository.NewGormProjectMemberRepository(db)
	projectRepo := repository.NewGormProjectRepository(db)
	projectService := service.NewProjectService(projectRepo, projectMemberRepo, userRepo)
	projectHandler := handler.NewProjectHandler(projectService)

	taskRepo := repository.NewGormTaskRepository(db)
	taskService := service.NewTaskService(taskRepo, projectRepo, projectMemberRepo, redisClient)
	taskHandler := handler.NewTaskHandler(taskService)
	commentRepo := repository.NewGormCommentRepository(db)
	commentHandler := handler.NewCommentHandler(taskService, commentRepo, userRepo, redisClient)

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.Logger(), middleware.Recovery())

	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	projectRoutes := router.Group("/projects", middleware.AuthMiddleware(jwtService))
	{
		projectRoutes.POST("", projectHandler.CreateProject)
		projectRoutes.GET("/:id", projectHandler.GetProjectByID)
		projectRoutes.PUT("/:id", projectHandler.UpdateProject)
		projectRoutes.DELETE("/:id", projectHandler.DeleteProject)
		projectRoutes.POST("/:id/tasks", taskHandler.CreateProjectTask)
	}

	taskRoutes := router.Group("/tasks", middleware.AuthMiddleware(jwtService))
	{
		taskRoutes.POST("/:id/comments", commentHandler.CreateTaskComment)
	}

	return router, db, redisClient
}

func TestAPIE2E_RegisterLoginProjectTaskComment(t *testing.T) {
	router, db, redisClient := setupE2ERouter(t)

	var (
		authToken string
		projectID int64
		taskID    int64
		userID    int64
		commentID int64
	)

	testEmail := fmt.Sprintf("e2e_%d@example.com", time.Now().UnixNano())
	testPassword := "123456"

	t.Cleanup(func() {
		ctx := context.Background()
		if commentID > 0 {
			db.Exec("DELETE FROM comments WHERE id = ?", commentID)
		}
		if taskID > 0 {
			db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
		}
		if projectID > 0 {
			db.Exec("DELETE FROM project_members WHERE project_id = ?", projectID)
			db.Exec("DELETE FROM projects WHERE id = ?", projectID)
		}
		if userID > 0 {
			db.Exec("DELETE FROM users WHERE id = ?", userID)
		} else {
			db.Exec("DELETE FROM users WHERE email = ?", testEmail)
		}
		if redisClient != nil {
			if userID > 0 {
				_ = redisClient.Del(ctx, fmt.Sprintf("tasks:user:%d", userID)).Err()
			}
			if taskID > 0 {
				_ = redisClient.Del(ctx, fmt.Sprintf("task:%d", taskID)).Err()
			}
			_ = redisClient.Close()
		}
	})

	t.Run("1. Register", func(t *testing.T) {
		body := map[string]any{
			"email":     testEmail,
			"password":  testPassword,
			"full_name": "E2E User",
		}
		rr := performRequest(router, http.MethodPost, "/auth/register", body, "")
		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp apiResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		var data registerData
		err = json.Unmarshal(resp.Data, &data)
		assert.NoError(t, err)
		assert.Greater(t, data.ID, int64(0))
		userID = data.ID
	})

	t.Run("2. Login", func(t *testing.T) {
		body := map[string]any{
			"email":    testEmail,
			"password": testPassword,
		}
		rr := performRequest(router, http.MethodPost, "/auth/login", body, "")
		assert.Equal(t, http.StatusOK, rr.Code)

		var resp apiResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		var data loginData
		err = json.Unmarshal(resp.Data, &data)
		assert.NoError(t, err)
		assert.NotEmpty(t, data.AccessToken)
		authToken = data.AccessToken
	})

	t.Run("3. Create Project", func(t *testing.T) {
		body := map[string]any{
			"name":        "E2E Project",
			"description": "integration test project",
		}
		rr := performRequest(router, http.MethodPost, "/projects", body, authToken)
		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp apiResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		var data projectData
		err = json.Unmarshal(resp.Data, &data)
		assert.NoError(t, err)
		assert.Greater(t, data.ID, int64(0))
		projectID = data.ID
	})

	t.Run("4. Create Project Task", func(t *testing.T) {
		path := fmt.Sprintf("/projects/%d/tasks", projectID)
		body := map[string]any{
			"project_id":  projectID,
			"title":       "E2E Task",
			"description": "integration task",
			"status":      "todo",
		}
		rr := performRequest(router, http.MethodPost, path, body, authToken)
		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp apiResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		var data taskData
		err = json.Unmarshal(resp.Data, &data)
		assert.NoError(t, err)
		assert.Greater(t, data.ID, int64(0))
		taskID = data.ID
	})

	t.Run("5. Create Comment", func(t *testing.T) {
		path := fmt.Sprintf("/tasks/%d/comments", taskID)
		body := map[string]any{
			"content": "E2E comment content",
		}
		rr := performRequest(router, http.MethodPost, path, body, authToken)
		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp apiResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		var data commentData
		err = json.Unmarshal(resp.Data, &data)
		assert.NoError(t, err)
		assert.Greater(t, data.ID, int64(0))
		assert.Equal(t, taskID, data.TaskID)
		assert.Equal(t, "E2E comment content", data.Content)
		commentID = data.ID
	})
}
