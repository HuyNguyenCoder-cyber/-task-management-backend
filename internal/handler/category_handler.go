package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/service"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	category, err := h.categoryService.CreateCategory(req.Name, req.Description)
	if err != nil {
		handleCategoryServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusCreated, "Category created successfully", toCategoryResponse(category))
}

func (h *CategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryService.ListCategories()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list categories", err)
		return
	}

	responses := make([]dto.CategoryResponse, 0, len(categories))

	for _, category := range categories {
		responses = append(responses, toCategoryResponse(category))
	}

	writeSuccess(c, http.StatusOK, "Categories retrieved successfully", responses)
}

func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	id := c.Param("id")

	category, err := h.categoryService.GetCategoryByID(id)
	if err != nil {
		handleCategoryServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Category retrieved successfully", toCategoryResponse(category))
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	category, err := h.categoryService.UpdateCategory(id, req.Name, req.Description)
	if err != nil {
		handleCategoryServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Category updated successfully", toCategoryResponse(category))
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	err := h.categoryService.DeleteCategory(id)
	if err != nil {
		handleCategoryServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Category deleted successfully", nil)
}

func toCategoryResponse(category *domain.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:          category.CatId,
		Name:        category.CatName,
		Description: category.Description,
		CreatedAt:   category.CreateAt,
		UpdatedAt:   category.UpdateAt,
	}
}

func handleCategoryServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCategoryNameRequired):
		writeError(c, http.StatusBadRequest, "Category name is required", err)

	case errors.Is(err, repository.ErrCategoryNotFound):
		writeError(c, http.StatusNotFound, "Category not found", err)

	default:
		writeError(c, http.StatusInternalServerError, "Internal server error", err)
	}
}
