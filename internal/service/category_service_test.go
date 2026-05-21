package service

import (
	"errors"
	"testing"
	"time"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/repository"
)

type mockCategoryRepository struct {
	categories map[string]*domain.Category
}

func newMockCategoryRepository() *mockCategoryRepository {
	return &mockCategoryRepository{
		categories: make(map[string]*domain.Category),
	}
}

func (m *mockCategoryRepository) Create(category *domain.Category) error {
	m.categories[category.CatId] = category
	return nil
}

func (m *mockCategoryRepository) GetByID(id string) (*domain.Category, error) {
	category, exists := m.categories[id]
	if !exists {
		return nil, repository.ErrCategoryNotFound
	}

	return category, nil
}

func (m *mockCategoryRepository) List() ([]*domain.Category, error) {
	categories := make([]*domain.Category, 0, len(m.categories))

	for _, category := range m.categories {
		categories = append(categories, category)
	}

	return categories, nil
}

func (m *mockCategoryRepository) Update(category *domain.Category) error {
	_, exists := m.categories[category.CatId]
	if !exists {
		return repository.ErrCategoryNotFound
	}

	m.categories[category.CatId] = category
	return nil
}

func (m *mockCategoryRepository) Delete(id string) error {
	_, exists := m.categories[id]
	if !exists {
		return repository.ErrCategoryNotFound
	}

	delete(m.categories, id)
	return nil
}

func TestCategoryService_CreateCategory_Success(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	category, err := categoryService.CreateCategory(
		"Backend",
		"Tasks related to backend development",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if category == nil {
		t.Fatal("expected category, got nil")
	}

	if category.CatId == "" {
		t.Error("expected category ID to be generated")
	}

	if category.CatId != "Backend" {
		t.Errorf("expected name Backend, got %s", category.CatName)
	}

	if category.Description != "Tasks related to backend development" {
		t.Errorf("expected description, got %s", category.Description)
	}

	if category.CreateAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if category.UpdateAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestCategoryService_CreateCategory_NameRequired(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	category, err := categoryService.CreateCategory(
		"   ",
		"Invalid category",
	)

	if category != nil {
		t.Errorf("expected category nil, got %v", category)
	}

	if !errors.Is(err, ErrCategoryNameRequired) {
		t.Errorf("expected ErrCategoryNameRequired, got %v", err)
	}
}

func TestCategoryService_ListCategories(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	_, err := categoryService.CreateCategory("Backend", "Backend tasks")
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	_, err = categoryService.CreateCategory("Frontend", "Frontend tasks")
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	categories, err := categoryService.ListCategories()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(categories))
	}
}

func TestCategoryService_GetCategoryByID_Success(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	createdCategory, err := categoryService.CreateCategory(
		"Bug",
		"Bug fixing tasks",
	)
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	foundCategory, err := categoryService.GetCategoryByID(createdCategory.CatId)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if foundCategory.CatId != createdCategory.CatId {
		t.Errorf("expected ID %s, got %s", createdCategory.CatId, foundCategory.CatId)
	}

	if foundCategory.CatName != createdCategory.CatName {
		t.Errorf("expected name %s, got %s", createdCategory.CatName, foundCategory.CatName)
	}
}

func TestCategoryService_GetCategoryByID_NotFound(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	category, err := categoryService.GetCategoryByID("not-found-id")

	if category != nil {
		t.Errorf("expected category nil, got %v", category)
	}

	if !errors.Is(err, repository.ErrCategoryNotFound) {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCategoryService_UpdateCategory_Success(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	now := time.Now()

	existingCategory := &domain.Category{
		CatId:       "category-1",
		CatName:     "Old Category",
		Description: "Old description",
		CreateAt:    now,
		UpdateAt:    now,
	}

	err := mockRepo.Create(existingCategory)
	if err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	oldUpdatedAt := existingCategory.UpdateAt

	updatedCategory, err := categoryService.UpdateCategory(
		"category-1",
		"New Category",
		"New description",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedCategory.CatId != existingCategory.CatId {
		t.Errorf("expected ID %s, got %s", existingCategory.CatId, updatedCategory.CatId)
	}

	if updatedCategory.CatName != "New Category" {
		t.Errorf("expected name New Category, got %s", updatedCategory.CatName)
	}

	if updatedCategory.Description != "New description" {
		t.Errorf("expected description New description, got %s", updatedCategory.Description)
	}

	if !updatedCategory.CreateAt.Equal(existingCategory.CreateAt) {
		t.Error("expected CreatedAt to stay unchanged")
	}

	if !updatedCategory.UpdateAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestCategoryService_UpdateCategory_NameRequired(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	category, err := categoryService.UpdateCategory(
		"category-1",
		"   ",
		"Invalid name",
	)

	if category != nil {
		t.Errorf("expected category nil, got %v", category)
	}

	if !errors.Is(err, ErrCategoryNameRequired) {
		t.Errorf("expected ErrCategoryNameRequired, got %v", err)
	}
}

func TestCategoryService_UpdateCategory_NotFound(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	category, err := categoryService.UpdateCategory(
		"not-found-id",
		"Valid Name",
		"Valid description",
	)

	if category != nil {
		t.Errorf("expected category nil, got %v", category)
	}

	if !errors.Is(err, repository.ErrCategoryNotFound) {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCategoryService_DeleteCategory_Success(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	category, err := categoryService.CreateCategory(
		"Learning",
		"Learning tasks",
	)
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	err = categoryService.DeleteCategory(category.CatId)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	deletedCategory, err := categoryService.GetCategoryByID(category.CatId)

	if deletedCategory != nil {
		t.Errorf("expected deleted category nil, got %v", deletedCategory)
	}

	if !errors.Is(err, repository.ErrCategoryNotFound) {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCategoryService_DeleteCategory_NotFound(t *testing.T) {
	mockRepo := newMockCategoryRepository()
	categoryService := NewCategoryService(mockRepo)

	err := categoryService.DeleteCategory("not-found-id")

	if !errors.Is(err, repository.ErrCategoryNotFound) {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}
