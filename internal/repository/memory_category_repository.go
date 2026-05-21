package repository

import (
	"errors"
	"task-management-backend/internal/domain"
)

var ErrCategoryNotFound = errors.New("category not found")

type MemoryCategoryRepository struct {
	categories map[string]*domain.Category
}

func NewMemoryCategoryRepository() *MemoryCategoryRepository {
	return &MemoryCategoryRepository{
		categories: make(map[string]*domain.Category),
	}
}
func (r *MemoryCategoryRepository) Create(category *domain.Category) error {
	r.categories[category.CatId] = category
	return nil
}
func (r *MemoryCategoryRepository) GetByID(id string) (*domain.Category, error) {
	category, exists := r.categories[id]
	if !exists {
		return nil, ErrCategoryNotFound
	}
	return category, nil
}
func (r *MemoryCategoryRepository) List() ([]*domain.Category, error) {
	categories := make([]*domain.Category, 0, len(r.categories))

	for _, category := range r.categories {
		categories = append(categories, category)
	}

	return categories, nil
}
func (r *MemoryCategoryRepository) Update(category *domain.Category) error {
	_, exists := r.categories[category.CatId]
	if !exists {
		return ErrCategoryNotFound
	}
	r.categories[category.CatId] = category
	return nil
}
func (r *MemoryCategoryRepository) Delete(id string) error {
	_, exists := r.categories[id]
	if !exists {
		return ErrCategoryNotFound
	}
	delete(r.categories, id)
	return nil
}

// biến check xem implement interface đủ chưa
var _ CategoryRepository = (*MemoryCategoryRepository)(nil)
