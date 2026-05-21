package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/repository"
)

var (
	ErrTaskTitleRequired = errors.New("task title is required")
	ErrInvalidTaskStatus = errors.New("invalid task status")
)

type taskService struct {
	taskRepo repository.TaskRepository
}

func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{
		taskRepo: taskRepo,
	}
}

func (s *taskService) CreateTask(title string, description string, status domain.TaskStatus, assignee string) (*domain.Task, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrTaskTitleRequired
	}

	if status == "" {
		status = domain.TaskStatusTodo
	}

	if !isValidTaskStatus(status) {
		return nil, ErrInvalidTaskStatus
	}

	now := time.Now()

	task := &domain.Task{
		ID:          generateTaskID(),
		Title:       title,
		Description: description,
		Status:      status,
		Assignee:    assignee,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.taskRepo.Create(task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) GetTaskByID(id string) (*domain.Task, error) {
	return s.taskRepo.GetByID(id)
}

func (s *taskService) ListTasks() ([]*domain.Task, error) {
	return s.taskRepo.List()
}

func (s *taskService) UpdateTask(id string, title string, description string, status domain.TaskStatus, assignee string) (*domain.Task, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrTaskTitleRequired
	}

	if !isValidTaskStatus(status) {
		return nil, ErrInvalidTaskStatus
	}

	existingTask, err := s.taskRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	updatedTask := &domain.Task{
		ID:          existingTask.ID,
		Title:       title,
		Description: description,
		Status:      status,
		Assignee:    assignee,
		CreatedAt:   existingTask.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	err = s.taskRepo.Update(updatedTask)
	if err != nil {
		return nil, err
	}

	return updatedTask, nil
}

func (s *taskService) DeleteTask(id string) error {
	return s.taskRepo.Delete(id)
}

func isValidTaskStatus(status domain.TaskStatus) bool {
	switch status {
	case domain.TaskStatusTodo,
		domain.TaskStatusInProgress,
		domain.TaskStatusDone:
		return true
	default:
		return false
	}
}

func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}