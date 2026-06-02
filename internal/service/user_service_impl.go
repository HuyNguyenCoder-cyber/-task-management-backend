package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"task-management-backend/internal/domain"
	"task-management-backend/internal/repository"
)

var (
	ErrUserEmailRequired        = errors.New("user email is required")
	ErrUserPasswordHashRequired = errors.New("user password hash is required")
	ErrUserFullNameRequired     = errors.New("user full name is required")
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.List(ctx)
}

func (s *userService) UpdateUser(ctx context.Context, id int64, email string, fullName string) (*domain.User, error) {
	email = strings.TrimSpace(email)
	fullName = strings.TrimSpace(fullName)

	if email == "" {
		return nil, ErrUserEmailRequired
	}
	if fullName == "" {
		return nil, ErrUserFullNameRequired
	}

	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updatedUser := &domain.User{
		ID:           existingUser.ID,
		Email:        email,
		PasswordHash: existingUser.PasswordHash,
		FullName:     fullName,
		CreatedAt:    existingUser.CreatedAt,
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Update(ctx, updatedUser); err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	return s.userRepo.Delete(ctx, id)
}
