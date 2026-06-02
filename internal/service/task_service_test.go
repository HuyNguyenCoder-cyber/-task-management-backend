package service

import (
	"context"
	"testing"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
)

type mockTaskRepo struct {
	createFn           func(context.Context, *domain.Task) error
	findByIDFn         func(context.Context, int64) (*domain.Task, error)
	findTasksForUserFn func(context.Context, int64) ([]*domain.Task, error)
	findByProjectIDFn  func(context.Context, int64) ([]*domain.Task, error)
	updateFn           func(context.Context, *domain.Task) error
	deleteFn           func(context.Context, int64) error
}

func (m *mockTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	if m.createFn != nil {
		return m.createFn(ctx, task)
	}
	return nil
}
func (m *mockTaskRepo) FindByID(ctx context.Context, id int64) (*domain.Task, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, repository.ErrTaskNotFound
}
func (m *mockTaskRepo) FindTasksForUser(ctx context.Context, userID int64) ([]*domain.Task, error) {
	if m.findTasksForUserFn != nil {
		return m.findTasksForUserFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockTaskRepo) FindByProjectID(ctx context.Context, projectID int64) ([]*domain.Task, error) {
	if m.findByProjectIDFn != nil {
		return m.findByProjectIDFn(ctx, projectID)
	}
	return nil, nil
}
func (m *mockTaskRepo) Update(ctx context.Context, task *domain.Task) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, task)
	}
	return nil
}
func (m *mockTaskRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestTaskService_CreateTask_Success(t *testing.T) {
	repo := &mockTaskRepo{createFn: func(ctx context.Context, task *domain.Task) error {
		task.ID = 100
		return nil
	}}
	svc := NewTaskService(repo, &mockProjectRepo{}, &mockProjectMemberRepo{}, nil)

	task, err := svc.CreateTask(context.Background(), 1, dto.CreateTaskRequest{Title: "  Write test  "})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.ID != 100 || task.Title != "Write test" || task.Status != domain.TaskStatusTodo {
		t.Fatalf("unexpected task: %#v", task)
	}
	if task.AssigneeID == nil || *task.AssigneeID != 1 {
		t.Fatalf("expected assignee to default to creator")
	}
}

func TestTaskService_CreateTask_ValidationError(t *testing.T) {
	svc := NewTaskService(&mockTaskRepo{}, &mockProjectRepo{}, &mockProjectMemberRepo{}, nil)
	_, err := svc.CreateTask(context.Background(), 1, dto.CreateTaskRequest{Title: "ok", Status: "invalid"})
	if err != ErrInvalidTaskStatus {
		t.Fatalf("expected ErrInvalidTaskStatus, got %v", err)
	}
}

func TestTaskService_CreateTask_BusinessRuleViolation_AssigneeOtherUser(t *testing.T) {
	assigneeID := int64(2)
	svc := NewTaskService(&mockTaskRepo{}, &mockProjectRepo{}, &mockProjectMemberRepo{}, nil)

	_, err := svc.CreateTask(context.Background(), 1, dto.CreateTaskRequest{Title: "task", AssigneeID: &assigneeID})
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskService_UpdateTask_BusinessRuleViolation_AssigneeCannotEditTitle(t *testing.T) {
	assigneeID := int64(2)
	repo := &mockTaskRepo{findByIDFn: func(ctx context.Context, id int64) (*domain.Task, error) {
		return &domain.Task{ID: id, CreatedBy: 1, AssigneeID: &assigneeID, Title: "old", Status: domain.TaskStatusTodo}, nil
	}}
	svc := NewTaskService(repo, &mockProjectRepo{}, &mockProjectMemberRepo{}, nil)

	newTitle := "new"
	_, err := svc.UpdateTask(context.Background(), 2, 10, dto.UpdateTaskRequest{TitleSet: true, Title: &newTitle})
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskService_UpdateTask_Success_AssigneeChangeStatus(t *testing.T) {
	assigneeID := int64(2)
	repo := &mockTaskRepo{findByIDFn: func(ctx context.Context, id int64) (*domain.Task, error) {
		return &domain.Task{ID: id, CreatedBy: 1, AssigneeID: &assigneeID, Status: domain.TaskStatusTodo}, nil
	}}
	svc := NewTaskService(repo, &mockProjectRepo{}, &mockProjectMemberRepo{}, nil)

	status := "done"
	task, err := svc.UpdateTask(context.Background(), 2, 10, dto.UpdateTaskRequest{StatusSet: true, Status: &status})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.Status != domain.TaskStatusDone {
		t.Fatalf("expected status done, got %s", task.Status)
	}
}
