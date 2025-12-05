package unit

import (
	"testing"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInMemoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	// Creating a simple sqlite-compatible users table manually.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	);
	`
	if execErr := db.Exec(createTableSQL).Error; execErr != nil {
		t.Fatalf("failed to create users table: %v", execErr)
	}

	return db
}

func TestUserRepository_Create_GetByEmail_GetByID_EmailExists_Update_Delete_GetAll(t *testing.T) {
	db := setupInMemoryDB(t)
	repo := repository.NewUserRepository(db)

	u := &domain.User{
		ID:           uuid.NewString(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
		Role:         "user",
	}

	// Create
	if err := repo.Create(u); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// GetByEmail
	gotByEmail, err := repo.GetByEmail(u.Email)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if gotByEmail.Email != u.Email {
		t.Fatalf("expected email %q, got %q", u.Email, gotByEmail.Email)
	}

	// GetByID
	gotByID, err := repo.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if gotByID.ID != u.ID {
		t.Fatalf("expected id %q, got %q", u.ID, gotByID.ID)
	}

	// EmailExists
	exists, err := repo.EmailExists(u.Email)
	if err != nil {
		t.Fatalf("EmailExists failed: %v", err)
	}
	if !exists {
		t.Fatalf("expected email to exist")
	}

	// Update (no changes — just check if there is no error)
	if err := repo.Update(gotByID); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// GetAll
	users, err := repo.GetAll(10, 0)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("expected at least 1 user from GetAll")
	}

	// UpdateRole
	if err := repo.UpdateRole(u.ID, "admin"); err != nil {
		t.Fatalf("UpdateRole failed: %v", err)
	}

	// Delete
	if err := repo.Delete(u.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// EmailExists should be false after delete
	existsAfter, err := repo.EmailExists(u.Email)
	if err != nil {
		t.Fatalf("EmailExists after delete failed: %v", err)
	}
	if existsAfter {
		t.Fatalf("expected email to NOT exist after delete")
	}
}
