package service

import (
	"errors"
	"testing"
	"time"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/repository"
)

type mockTaskRepository struct {
	tasks map[string]*domain.Task

	createErr  error
	getByIDErr error
	updateErr  error
	deleteErr  error
}

func newMockTaskRepository() *mockTaskRepository {
	return &mockTaskRepository{
		tasks: make(map[string]*domain.Task),
	}
}

func (m *mockTaskRepository) Create(task *domain.Task) error {
	if m.createErr != nil {
		return m.createErr
	}

	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) GetByID(id string) (*domain.Task, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	task, exists := m.tasks[id]
	if !exists {
		return nil, repository.ErrTaskNotFound
	}

	return task, nil
}

func (m *mockTaskRepository) List() ([]*domain.Task, error) {
	tasks := make([]*domain.Task, 0, len(m.tasks))

	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *mockTaskRepository) Update(task *domain.Task) error {
	if m.updateErr != nil {
		return m.updateErr
	}

	_, exists := m.tasks[task.ID]
	if !exists {
		return repository.ErrTaskNotFound
	}

	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}

	_, exists := m.tasks[id]
	if !exists {
		return repository.ErrTaskNotFound
	}

	delete(m.tasks, id)
	return nil
}

func TestTaskService_CreateTask_Success(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	task, err := taskService.CreateTask(
		"Learn service test",
		"Write unit test for service layer",
		domain.TaskStatusTodo,
		"Huy",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if task == nil {
		t.Fatal("expected task, got nil")
	}

	if task.ID == "" {
		t.Error("expected task ID to be generated")
	}

	if task.Title != "Learn service test" {
		t.Errorf("expected title %s, got %s", "Learn service test", task.Title)
	}

	if task.Description != "Write unit test for service layer" {
		t.Errorf("expected description %s, got %s", "Write unit test for service layer", task.Description)
	}

	if task.Status != domain.TaskStatusTodo {
		t.Errorf("expected status %s, got %s", domain.TaskStatusTodo, task.Status)
	}

	if task.Assignee != "Huy" {
		t.Errorf("expected assignee %s, got %s", "Huy", task.Assignee)
	}

	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if task.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}

	savedTask, err := mockRepo.GetByID(task.ID)
	if err != nil {
		t.Fatalf("expected task saved in repo, got error %v", err)
	}

	if savedTask.ID != task.ID {
		t.Errorf("expected saved task ID %s, got %s", task.ID, savedTask.ID)
	}
}

func TestTaskService_CreateTask_DefaultStatus(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	task, err := taskService.CreateTask(
		"Task without status",
		"Status should default to todo",
		"",
		"Huy",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if task.Status != domain.TaskStatusTodo {
		t.Errorf("expected default status %s, got %s", domain.TaskStatusTodo, task.Status)
	}
}

func TestTaskService_CreateTask_TitleRequired(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	task, err := taskService.CreateTask(
		"   ",
		"Invalid empty title",
		domain.TaskStatusTodo,
		"Huy",
	)

	if task != nil {
		t.Errorf("expected task nil, got %v", task)
	}

	if !errors.Is(err, ErrTaskTitleRequired) {
		t.Errorf("expected ErrTaskTitleRequired, got %v", err)
	}
}

func TestTaskService_CreateTask_InvalidStatus(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	task, err := taskService.CreateTask(
		"Invalid status task",
		"Status is invalid",
		domain.TaskStatus("wrong_status"),
		"Huy",
	)

	if task != nil {
		t.Errorf("expected task nil, got %v", task)
	}

	if !errors.Is(err, ErrInvalidTaskStatus) {
		t.Errorf("expected ErrInvalidTaskStatus, got %v", err)
	}
}

func TestTaskService_UpdateTask_Success(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	now := time.Now()

	existingTask := &domain.Task{
		ID:          "task-1",
		Title:       "Old title",
		Description: "Old description",
		Status:      domain.TaskStatusTodo,
		Assignee:    "Huy",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := mockRepo.Create(existingTask)
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	updatedTask, err := taskService.UpdateTask(
		"task-1",
		"New title",
		"New description",
		domain.TaskStatusInProgress,
		"Huy Nguyen",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedTask.ID != existingTask.ID {
		t.Errorf("expected ID %s, got %s", existingTask.ID, updatedTask.ID)
	}

	if updatedTask.Title != "New title" {
		t.Errorf("expected title %s, got %s", "New title", updatedTask.Title)
	}

	if updatedTask.Description != "New description" {
		t.Errorf("expected description %s, got %s", "New description", updatedTask.Description)
	}

	if updatedTask.Status != domain.TaskStatusInProgress {
		t.Errorf("expected status %s, got %s", domain.TaskStatusInProgress, updatedTask.Status)
	}

	if updatedTask.Assignee != "Huy Nguyen" {
		t.Errorf("expected assignee %s, got %s", "Huy Nguyen", updatedTask.Assignee)
	}

	if !updatedTask.CreatedAt.Equal(existingTask.CreatedAt) {
		t.Error("expected CreatedAt to stay unchanged")
	}

	if !updatedTask.UpdatedAt.After(existingTask.UpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestTaskService_UpdateTask_TitleRequired(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	updatedTask, err := taskService.UpdateTask(
		"task-1",
		"   ",
		"Invalid title",
		domain.TaskStatusTodo,
		"Huy",
	)

	if updatedTask != nil {
		t.Errorf("expected updatedTask nil, got %v", updatedTask)
	}

	if !errors.Is(err, ErrTaskTitleRequired) {
		t.Errorf("expected ErrTaskTitleRequired, got %v", err)
	}
}

func TestTaskService_UpdateTask_InvalidStatus(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	updatedTask, err := taskService.UpdateTask(
		"task-1",
		"Valid title",
		"Invalid status",
		domain.TaskStatus("wrong_status"),
		"Huy",
	)

	if updatedTask != nil {
		t.Errorf("expected updatedTask nil, got %v", updatedTask)
	}

	if !errors.Is(err, ErrInvalidTaskStatus) {
		t.Errorf("expected ErrInvalidTaskStatus, got %v", err)
	}
}

func TestTaskService_UpdateTask_NotFound(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	updatedTask, err := taskService.UpdateTask(
		"not-found-id",
		"Valid title",
		"Valid description",
		domain.TaskStatusTodo,
		"Huy",
	)

	if updatedTask != nil {
		t.Errorf("expected updatedTask nil, got %v", updatedTask)
	}

	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskService_DeleteTask_Success(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	now := time.Now()

	task := &domain.Task{
		ID:          "task-1",
		Title:       "Task to delete",
		Description: "This task should be deleted",
		Status:      domain.TaskStatusDone,
		Assignee:    "Huy",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := mockRepo.Create(task)
	if err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	err = taskService.DeleteTask("task-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	deletedTask, err := mockRepo.GetByID("task-1")

	if deletedTask != nil {
		t.Errorf("expected deleted task nil, got %v", deletedTask)
	}

	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound after delete, got %v", err)
	}
}

func TestTaskService_DeleteTask_NotFound(t *testing.T) {
	mockRepo := newMockTaskRepository()
	taskService := NewTaskService(mockRepo)

	err := taskService.DeleteTask("not-found-id")

	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}
