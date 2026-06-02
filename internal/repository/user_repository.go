package repository

import (
	"context"
	"errors"

	"task-management-backend/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
}
