package unit

import (
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
)

// Using the already existing setupAuthRepo(t *testing.T) from setup.go,
// it returns *repository.UserRepository (in-memory sqlite).
func setupAuthServiceWithRealRepo(t *testing.T) *service.AuthService {
	repo := setupAuthRepo(t) // из setup.go
	jwtSecret := "test-secret"
	ttl := time.Hour
	return service.NewAuthService(repo, jwtSecret, ttl)
}

func TestAuthService_Register(t *testing.T) {
	svc := setupAuthServiceWithRealRepo(t)

	t.Run("successful registration", func(t *testing.T) {
		req := &domain.CreateUserRequest{
			Email:    "newuser@example.com",
			Password: "password123",
			Name:     "New User",
		}

		user, err := svc.Register(req)
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		if user.Email != req.Email {
			t.Errorf("expected email %q, got %q", req.Email, user.Email)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		req := &domain.CreateUserRequest{
			Email:    "newuser@example.com",
			Password: "password123",
			Name:     "New User",
		}

		_, err := svc.Register(req)
		if err == nil {
			t.Error("expected error for duplicate email, got nil")
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	svc := setupAuthServiceWithRealRepo(t)

	// First, register the user
	email := "loginuser@example.com"
	password := "password123"
	_, _ = svc.Register(&domain.CreateUserRequest{
		Email:    email,
		Password: password,
		Name:     "Login User",
	})

	t.Run("successful login", func(t *testing.T) {
		resp, err := svc.Login(&domain.LoginRequest{
			Email:    email,
			Password: password,
		})
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}
		if resp.Token == "" {
			t.Error("expected non-empty token")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login(&domain.LoginRequest{
			Email:    email,
			Password: "wrongpass",
		})
		if err == nil {
			t.Error("expected error for wrong password, got nil")
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := svc.Login(&domain.LoginRequest{
			Email:    "nouser@example.com",
			Password: "any",
		})
		if err == nil {
			t.Error("expected error for non-existent user, got nil")
		}
	})
}
