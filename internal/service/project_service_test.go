package service

import (
	"context"
	"testing"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
)

type mockProjectRepo struct {
	createFn      func(context.Context, *domain.Project) error
	getByIDFn     func(context.Context, int64) (*domain.Project, error)
	listForUserFn func(context.Context, int64) ([]*domain.Project, error)
	updateFn      func(context.Context, *domain.Project) error
	deleteFn      func(context.Context, int64) error
}

func (m *mockProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	return nil
}
func (m *mockProjectRepo) GetByID(ctx context.Context, id int64) (*domain.Project, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, repository.ErrProjectNotFound
}
func (m *mockProjectRepo) ListForUser(ctx context.Context, userID int64) ([]*domain.Project, error) {
	if m.listForUserFn != nil {
		return m.listForUserFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return nil
}
func (m *mockProjectRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockProjectMemberRepo struct {
	createFn           func(context.Context, *domain.ProjectMember) error
	getByProjectUserFn func(context.Context, int64, int64) (*domain.ProjectMember, error)
	listByProjectIDFn  func(context.Context, int64) ([]*domain.ProjectMemberInfo, error)
}

func (m *mockProjectMemberRepo) Create(ctx context.Context, member *domain.ProjectMember) error {
	if m.createFn != nil {
		return m.createFn(ctx, member)
	}
	return nil
}
func (m *mockProjectMemberRepo) GetByProjectAndUser(ctx context.Context, projectID int64, userID int64) (*domain.ProjectMember, error) {
	if m.getByProjectUserFn != nil {
		return m.getByProjectUserFn(ctx, projectID, userID)
	}
	return nil, repository.ErrProjectMemberNotFound
}
func (m *mockProjectMemberRepo) ListByProjectID(ctx context.Context, projectID int64) ([]*domain.ProjectMemberInfo, error) {
	if m.listByProjectIDFn != nil {
		return m.listByProjectIDFn(ctx, projectID)
	}
	return nil, nil
}

func TestProjectService_CreateProject_Success(t *testing.T) {
	projectRepo := &mockProjectRepo{createFn: func(ctx context.Context, p *domain.Project) error {
		p.ID = 11
		return nil
	}}
	memberRepo := &mockProjectMemberRepo{}
	svc := NewProjectService(projectRepo, memberRepo, &mockUserRepo{})

	p, err := svc.CreateProject(context.Background(), 1, "  Project A  ", "desc")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.ID != 11 || p.Name != "Project A" {
		t.Fatalf("unexpected project result: %#v", p)
	}
}

func TestProjectService_CreateProject_ValidationError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{}, &mockProjectMemberRepo{}, &mockUserRepo{})
	_, err := svc.CreateProject(context.Background(), 1, "   ", "desc")
	if err != ErrProjectNameRequired {
		t.Fatalf("expected ErrProjectNameRequired, got %v", err)
	}
}

func TestProjectService_AddProjectMember_BusinessRuleViolation_AlreadyMember(t *testing.T) {
	projectRepo := &mockProjectRepo{getByIDFn: func(ctx context.Context, id int64) (*domain.Project, error) {
		return &domain.Project{ID: id, OwnerID: 1}, nil
	}}
	memberRepo := &mockProjectMemberRepo{
		getByProjectUserFn: func(ctx context.Context, projectID int64, userID int64) (*domain.ProjectMember, error) {
			if userID == 2 {
				return &domain.ProjectMember{ProjectID: projectID, UserID: userID}, nil
			}
			return &domain.ProjectMember{ProjectID: projectID, UserID: userID}, nil
		},
	}
	userRepo := &mockUserRepo{findByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
		return &domain.User{ID: 2, Email: email, FullName: "Member"}, nil
	}}
	svc := NewProjectService(projectRepo, memberRepo, userRepo)

	_, err := svc.AddProjectMember(context.Background(), 1, 9, "member@example.com")
	if err != ErrProjectMemberExists {
		t.Fatalf("expected ErrProjectMemberExists, got %v", err)
	}
}

func TestProjectService_ListProjectMembers_Success(t *testing.T) {
	projectRepo := &mockProjectRepo{getByIDFn: func(ctx context.Context, id int64) (*domain.Project, error) {
		return &domain.Project{ID: id, OwnerID: 1}, nil
	}}
	memberRepo := &mockProjectMemberRepo{
		getByProjectUserFn: func(ctx context.Context, projectID int64, userID int64) (*domain.ProjectMember, error) {
			return &domain.ProjectMember{ProjectID: projectID, UserID: userID, Role: domain.ProjectMemberRoleOwner}, nil
		},
		listByProjectIDFn: func(ctx context.Context, projectID int64) ([]*domain.ProjectMemberInfo, error) {
			return []*domain.ProjectMemberInfo{{ID: 1, Email: "a@example.com", FullName: "A", Role: domain.ProjectMemberRoleOwner}}, nil
		},
	}
	svc := NewProjectService(projectRepo, memberRepo, &mockUserRepo{})

	res, err := svc.ListProjectMembers(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res) != 1 || res[0].Email != "a@example.com" {
		t.Fatalf("unexpected members: %#v", res)
	}
}

var _ = dto.ProjectMemberResponse{}
