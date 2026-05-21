package service

import "task-management-backend/internal/domain"

type TaskService interface {
	CreateTask(title string, description string, status domain.TaskStatus, assignee string) (*domain.Task, error)
	GetTaskByID(id string) (*domain.Task, error)
	ListTasks() ([]*domain.Task, error)
	UpdateTask(id string, title string, description string, status domain.TaskStatus, assignee string) (*domain.Task, error)
	DeleteTask(id string) error
}