package service

import "task-management-backend/internal/domain"

type CategoryService interface {
	CreateCategory(name string, description string) (*domain.Category, error)
	GetCategoryByID(id string) (*domain.Category, error)
	ListCategories() ([]*domain.Category, error)
	UpdateCategory(id string, name string, description string) (*domain.Category, error)
	DeleteCategory(id string) error
}
