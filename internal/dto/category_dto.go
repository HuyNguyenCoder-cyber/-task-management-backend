package dto

import "time"

type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type CategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
