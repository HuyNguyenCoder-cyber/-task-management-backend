package dto

import "time"

type UpdateUserRequest struct {
	Email    string `json:"email" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
