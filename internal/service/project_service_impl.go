package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
)

var (
	ErrProjectNameRequired        = errors.New("project name is required")
	ErrProjectOwnerRequired       = errors.New("project owner is required")
	ErrUnauthorized               = errors.New("unauthorized")
	ErrForbidden                  = errors.New("forbidden")
	ErrProjectMemberExists        = errors.New("project member already exists")
	ErrProjectMemberEmailRequired = errors.New("project member email is required")
)

type projectService struct {
	projectRepo       repository.ProjectRepository
	projectMemberRepo repository.ProjectMemberRepository
	userRepo          repository.UserRepository
}

func NewProjectService(projectRepo repository.ProjectRepository, projectMemberRepo repository.ProjectMemberRepository, userRepo repository.UserRepository) ProjectService {
	return &projectService{
		projectRepo:       projectRepo,
		projectMemberRepo: projectMemberRepo,
		userRepo:          userRepo,
	}
}

func (s *projectService) CreateProject(ctx context.Context, userID int64, name string, description string) (*domain.Project, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrProjectNameRequired
	}
	if userID == 0 {
		return nil, ErrProjectOwnerRequired
	}

	now := time.Now()
	project := &domain.Project{
		Name:        name,
		Description: description,
		OwnerID:     userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	member := &domain.ProjectMember{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      domain.ProjectMemberRoleOwner,
		CreatedAt: now,
	}
	if err := s.projectMemberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) GetProjectByID(ctx context.Context, userID int64, id int64) (*domain.Project, error) {
	if userID == 0 {
		return nil, ErrProjectOwnerRequired
	}

	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.ensureProjectMember(ctx, userID, id); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) ListProjectsByUserID(ctx context.Context, userID int64) ([]*domain.Project, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}

	return s.projectRepo.ListForUser(ctx, userID)
}

func (s *projectService) UpdateProject(ctx context.Context, userID int64, id int64, name string, description string) (*domain.Project, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrProjectNameRequired
	}
	if userID == 0 {
		return nil, ErrProjectOwnerRequired
	}

	existingProject, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existingProject.OwnerID != userID {
		return nil, ErrForbidden
	}

	updatedProject := &domain.Project{
		ID:          existingProject.ID,
		Name:        name,
		Description: description,
		OwnerID:     existingProject.OwnerID,
		CreatedAt:   existingProject.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	if err := s.projectRepo.Update(ctx, updatedProject); err != nil {
		return nil, err
	}

	return updatedProject, nil
}

func (s *projectService) DeleteProject(ctx context.Context, userID int64, id int64) error {
	if userID == 0 {
		return ErrProjectOwnerRequired
	}

	existingProject, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existingProject.OwnerID != userID {
		return ErrForbidden
	}

	return s.projectRepo.Delete(ctx, id)
}

func (s *projectService) AddProjectMember(ctx context.Context, userID int64, projectID int64, email string) (*dto.ProjectMemberResponse, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, ErrProjectMemberEmailRequired
	}
	if err := s.ensureProjectOwner(ctx, userID, projectID); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	existingMember, err := s.projectMemberRepo.GetByProjectAndUser(ctx, projectID, user.ID)
	if err == nil && existingMember != nil {
		return nil, ErrProjectMemberExists
	}
	if err != nil && !errors.Is(err, repository.ErrProjectMemberNotFound) {
		return nil, err
	}

	member := &domain.ProjectMember{
		ProjectID: projectID,
		UserID:    user.ID,
		Role:      domain.ProjectMemberRoleMember,
		CreatedAt: time.Now(),
	}
	if err := s.projectMemberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return &dto.ProjectMemberResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     member.Role,
	}, nil
}

func (s *projectService) ListProjectMembers(ctx context.Context, userID int64, projectID int64) ([]dto.ProjectMemberResponse, error) {
	if _, err := s.ensureProjectMember(ctx, userID, projectID); err != nil {
		return nil, err
	}

	members, err := s.projectMemberRepo.ListByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProjectMemberResponse, 0, len(members))
	for _, member := range members {
		responses = append(responses, dto.ProjectMemberResponse{
			ID:       member.ID,
			Email:    member.Email,
			FullName: member.FullName,
			Role:     member.Role,
		})
	}

	return responses, nil
}

func (s *projectService) ensureProjectOwner(ctx context.Context, userID int64, projectID int64) error {
	if userID == 0 {
		return ErrUnauthorized
	}

	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project.OwnerID != userID {
		return ErrForbidden
	}

	return nil
}

func (s *projectService) ensureProjectMember(ctx context.Context, userID int64, projectID int64) (*domain.ProjectMember, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		return nil, err
	}

	member, err := s.projectMemberRepo.GetByProjectAndUser(ctx, projectID, userID)
	if errors.Is(err, repository.ErrProjectMemberNotFound) {
		return nil, ErrForbidden
	}
	if err != nil {
		return nil, err
	}

	return member, nil
}
