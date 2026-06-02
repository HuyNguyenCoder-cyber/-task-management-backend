package service

import (
	"context"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
)

type ProjectService interface {
	CreateProject(ctx context.Context, userID int64, name string, description string) (*domain.Project, error)
	GetProjectByID(ctx context.Context, userID int64, id int64) (*domain.Project, error)
	ListProjectsByUserID(ctx context.Context, userID int64) ([]*domain.Project, error)
	UpdateProject(ctx context.Context, userID int64, id int64, name string, description string) (*domain.Project, error)
	DeleteProject(ctx context.Context, userID int64, id int64) error
	AddProjectMember(ctx context.Context, userID int64, projectID int64, email string) (*dto.ProjectMemberResponse, error)
	ListProjectMembers(ctx context.Context, userID int64, projectID int64) ([]dto.ProjectMemberResponse, error)
}
