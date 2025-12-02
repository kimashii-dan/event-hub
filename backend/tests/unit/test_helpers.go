// backend/tests/unit/test_helpers.go
// Общие helper функции для unit и integration тестов
package unit

import (
	"testing"

	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupAuthRepo создаёт in-memory БД для тестирования auth
func SetupAuthRepo(t *testing.T) *repository.UserRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

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
	if execErr := db.Exec(createTableSQL).Error; execErr != nil {
		t.Fatalf("failed to create users table: %v", execErr)
	}

	return repository.NewUserRepository(db)
}

// SetupEventRepo создаёт in-memory БД для тестирования events
func SetupEventRepo(t *testing.T) *repository.EventRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		date DATETIME,
		location TEXT,
		organizer_id TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	);
	`
	if execErr := db.Exec(createTableSQL).Error; execErr != nil {
		t.Fatalf("failed to create events table: %v", execErr)
	}

	return repository.NewEventRepository(db)
}

// SetupTestDB создаёт универсальную in-memory БД
func SetupTestDB(t *testing.T, tables ...string) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	return db
}
