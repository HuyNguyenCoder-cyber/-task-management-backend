package service

import (
	"context"

	"task-management-backend/internal/domain"
)

type UserService interface {
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	ListUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUser(ctx context.Context, id int64, email string, fullName string) (*domain.User, error)
	DeleteUser(ctx context.Context, id int64) error
}
