package service

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"task-management-backend/internal/auth"
	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
)

type mockUserRepo struct {
	createFn      func(ctx context.Context, user *domain.User) error
	getByIDFn     func(ctx context.Context, id int64) (*domain.User, error)
	findByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	getByEmailFn  func(ctx context.Context, email string) (*domain.User, error)
	listFn        func(ctx context.Context) ([]*domain.User, error)
	updateFn      func(ctx context.Context, user *domain.User) error
	deleteFn      func(ctx context.Context, id int64) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepo) List(ctx context.Context) ([]*domain.User, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestAuthService_Register_Success(t *testing.T) {
	ctx := context.Background()

	repo := &mockUserRepo{
		findByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, repository.ErrUserNotFound
		},
		createFn: func(ctx context.Context, user *domain.User) error {
			user.ID = 101
			return nil
		},
	}

	svc := NewAuthService(repo, auth.NewJWTService("secret", 1))
	resp, err := svc.Register(ctx, dto.RegisterRequest{
		Email:    "user@example.com",
		Password: "123456",
		FullName: "Test User",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil || resp.ID != 101 {
		t.Fatalf("expected created user id=101, got %#v", resp)
	}
}

func TestAuthService_Register_ValidationError(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{}, auth.NewJWTService("secret", 1))

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Email:    "user@example.com",
		Password: "123",
		FullName: "User",
	})

	if err != ErrAuthPasswordTooShort {
		t.Fatalf("expected ErrAuthPasswordTooShort, got %v", err)
	}
}

func TestAuthService_Register_BusinessRuleViolation_DuplicateEmail(t *testing.T) {
	repo := &mockUserRepo{
		findByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: 1, Email: email}, nil
		},
	}
	svc := NewAuthService(repo, auth.NewJWTService("secret", 1))

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Email:    "dup@example.com",
		Password: "123456",
		FullName: "Dup",
	})

	if err != ErrAuthEmailAlreadyExists {
		t.Fatalf("expected ErrAuthEmailAlreadyExists, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		findByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: 7, Email: email, FullName: "Login User", PasswordHash: string(hash)}, nil
		},
	}
	svc := NewAuthService(repo, auth.NewJWTService("secret", 1))

	resp, err := svc.Login(context.Background(), dto.LoginRequest{Email: "user@example.com", Password: "123456"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" || resp.User.ID != 7 {
		t.Fatalf("expected token and user id=7, got %#v", resp)
	}
}
