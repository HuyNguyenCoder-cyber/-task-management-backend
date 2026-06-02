package repository

import (
	"context"

	"task-management-backend/internal/domain"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	FindByTaskID(ctx context.Context, taskID int64) ([]*domain.Comment, error)
}

