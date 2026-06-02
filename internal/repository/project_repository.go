package repository

import (
	"context"
	"errors"

	"task-management-backend/internal/domain"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id int64) (*domain.Project, error)
	ListForUser(ctx context.Context, userID int64) ([]*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id int64) error
}
