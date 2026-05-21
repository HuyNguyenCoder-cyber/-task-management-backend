package repository

import "task-management-backend/internal/domain"

type CategoryRepository interface {
	Create(category *domain.Category) error
	GetByID(id string) (*domain.Category, error)
	List() ([]*domain.Category, error)
	Update(category *domain.Category) error
	Delete(id string) error
}
