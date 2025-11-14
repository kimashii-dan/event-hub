package unit

import (
	"fmt"
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func (m *mockUserRepo) Create(u *domain.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) GetByEmail(email string) (*domain.User, error) {
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockUserRepo) EmailExists(email string) (bool, error) {
	_, ok := m.users[email]
	return ok, nil
}

func TestRegister(t *testing.T) {
	mockRepo := &mockUserRepo{users: make(map[string]*domain.User)}
	auth := service.NewAuthService(mockRepo, "testsecret", time.Hour)

	req := &domain.CreateUserRequest{
		Email:    "test@mail.com",
		Password: "password123",
		Name:     "Test User",
	}

	user, err := auth.Register(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Email != "test@mail.com" {
		t.Fatalf("email mismatch")
	}
}
