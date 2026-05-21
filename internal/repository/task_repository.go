package repository

import "task-management-backend/internal/domain"

type TaskRepository interface {
	Create(task *domain.Task) error
	GetByID(id string) (*domain.Task, error)
	List() ([]*domain.Task, error)
	Update(task *domain.Task) error
	Delete(id string) error
}
