package unit

import (
	"log"
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

// setupAuthRepo creates an in-memory SQLite DB and returns a UserRepository.
func setupAuthRepo(t *testing.T) *repository.UserRepository {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to sqlite in-memory DB: %v", err)
	}

	// Create the users table manually for SQLite compatibility.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	);
	`
	if err := db.Exec(createTableSQL).Error; err != nil {
		t.Fatalf("failed to create users table (sqlite): %v", err)
	}

	// Return a repository using this DB.
	repo := repository.NewUserRepository(db)
	return repo
}

// createTestUser creates and saves a test user in the repository.
func createTestUser(repo *repository.UserRepository, email, password, name string) *domain.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(user); err != nil {
		log.Fatalf("failed to create test user: %v", err)
	}
	return user
}
