package repository

import (
	"context"
	"errors"

	"task-management-backend/internal/domain"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	FindByID(ctx context.Context, id int64) (*domain.Task, error)
	FindTasksForUser(ctx context.Context, userID int64) ([]*domain.Task, error)
	FindByProjectID(ctx context.Context, projectID int64) ([]*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id int64) error
}
