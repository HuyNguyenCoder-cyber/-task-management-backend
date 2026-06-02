package repository

import (
	"context"
	"errors"

	"task-management-backend/internal/domain"
)

var ErrProjectMemberNotFound = errors.New("project member not found")

type ProjectMemberRepository interface {
	Create(ctx context.Context, member *domain.ProjectMember) error
	GetByProjectAndUser(ctx context.Context, projectID int64, userID int64) (*domain.ProjectMember, error)
	ListByProjectID(ctx context.Context, projectID int64) ([]*domain.ProjectMemberInfo, error)
}
