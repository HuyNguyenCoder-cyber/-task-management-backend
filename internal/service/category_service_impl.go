package service

import (
	"errors"
	"fmt"
	"strings"
	"task-management-backend/internal/domain"
	"task-management-backend/internal/repository"
	"time"
)

var ErrCategoryNameRequired = errors.New("category name is required")

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}
func (s *categoryService) CreateCategory(name string, description string) (*domain.Category, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrCategoryNameRequired
	}

	now := time.Now()

	category := &domain.Category{
		CatId:       generateCategoryID(),
		CatName:     name,
		Description: description,
		CreateAt:    now,
		UpdateAt:    now,
	}

	err := s.categoryRepo.Create(category)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *categoryService) GetCategoryByID(id string) (*domain.Category, error) {
	return s.categoryRepo.GetByID(id)
}

func (s *categoryService) ListCategories() ([]*domain.Category, error) {
	return s.categoryRepo.List()
}

func (s *categoryService) UpdateCategory(id string, name string, description string) (*domain.Category, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrCategoryNameRequired
	}

	existingCategory, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	updatedCategory := &domain.Category{
		CatId:       existingCategory.CatId,
		CatName:     name,
		Description: description,
		CreateAt:    existingCategory.CreateAt,
		UpdateAt:    time.Now(),
	}

	err = s.categoryRepo.Update(updatedCategory)
	if err != nil {
		return nil, err
	}

	return updatedCategory, nil
}

func (s *categoryService) DeleteCategory(id string) error {
	return s.categoryRepo.Delete(id)
}
func generateCategoryID() string {
	return fmt.Sprintf("category-%d", time.Now().UnixNano())
}
