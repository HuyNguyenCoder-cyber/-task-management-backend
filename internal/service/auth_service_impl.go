package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"task-management-backend/internal/auth"
	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
)

var (
	ErrAuthEmailRequired      = errors.New("email is required")
	ErrAuthInvalidEmail       = errors.New("email format is invalid")
	ErrAuthPasswordRequired   = errors.New("password is required")
	ErrAuthPasswordTooShort   = errors.New("password must be at least 6 characters")
	ErrAuthFullNameRequired   = errors.New("full name is required")
	ErrAuthEmailAlreadyExists = errors.New("email already exists")
	ErrAuthInvalidCredentials = errors.New("invalid email or password")
)

type authService struct {
	userRepo   repository.UserRepository
	jwtService *auth.JWTService
}

func NewAuthService(userRepo repository.UserRepository, jwtService *auth.JWTService) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	req.FullName = strings.TrimSpace(req.FullName)

	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	_, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrAuthEmailAlreadyExists
	}
	// Hàm errors.Is(A, B) là gì?
	// Hàm này trong Go dùng để so sánh xem Lỗi A có bằng (hoặc cùng loại) với Lỗi B hay không. Nó chỉ trả về 2 kết quả:

	// Trả về true (Đúng) nếu hai lỗi giống nhau.
	// Trả về false (Sai) nếu hai lỗi khác nhau.
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &domain.User{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		FullName:     req.FullName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrAuthInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrAuthInvalidCredentials
	}

	accessToken, expiresIn, err := s.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User: dto.AuthUserInfo{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
		},
	}, nil
}

func validateRegisterRequest(req dto.RegisterRequest) error {
	if req.Email == "" {
		return ErrAuthEmailRequired
	}
	if !isValidEmail(req.Email) {
		return ErrAuthInvalidEmail
	}
	if req.Password == "" {
		return ErrAuthPasswordRequired
	}
	if len(req.Password) < 6 {
		return ErrAuthPasswordTooShort
	}
	if req.FullName == "" {
		return ErrAuthFullNameRequired
	}

	return nil
}

func validateLoginRequest(req dto.LoginRequest) error {
	if req.Email == "" {
		return ErrAuthEmailRequired
	}
	if req.Password == "" {
		return ErrAuthPasswordRequired
	}

	return nil
}

func isValidEmail(email string) bool {
	parsedAddress, err := mail.ParseAddress(email)
	return err == nil && parsedAddress.Address == email
}
