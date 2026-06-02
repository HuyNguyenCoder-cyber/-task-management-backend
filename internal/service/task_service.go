package service

import (
	"context"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
)

type TaskService interface {
	CreateTask(ctx context.Context, userID int64, req dto.CreateTaskRequest) (*domain.Task, error)
	CreateProjectTask(ctx context.Context, userID int64, projectID int64, req dto.CreateTaskRequest) (*domain.Task, error)
	GetTaskByID(ctx context.Context, userID int64, id int64) (*domain.Task, error)
	ListTasks(ctx context.Context, userID int64, projectID *int64) ([]*domain.Task, error)
	UpdateTask(ctx context.Context, userID int64, id int64, req dto.UpdateTaskRequest) (*domain.Task, error)
	DeleteTask(ctx context.Context, userID int64, id int64) error
}
