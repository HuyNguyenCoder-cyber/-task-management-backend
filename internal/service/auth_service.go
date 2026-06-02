package service

import (
	"context"

	"task-management-backend/internal/dto"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
}
