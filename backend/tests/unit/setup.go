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

// SetupAuthRepo создаёт in-memory SQLite DB и возвращает *repository.UserRepository
// Не трогаем продовой код — здесь создаём sqlite-совместимую таблицу вручную,
// чтобы избежать SQL с uuid_generate_v4(), который не поддерживает sqlite.
func setupAuthRepo(t *testing.T) *repository.UserRepository {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to sqlite in-memory DB: %v", err)
	}

	// Если используем sqlite — создаём таблицу users вручную (sqlite-совместимая схема).
	// Это предотвращает генерацию Postgres-специфичных выражений типа uuid_generate_v4().
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

	// Если в проекте есть дополнительные миграции/индексы, можно добавить их здесь.
	// Например: db.Exec("CREATE INDEX idx_users_email ON users(email);")

	// Теперь создаём реальный репозиторий, который использует это DB.
	repo := repository.NewUserRepository(db)
	return repo
}

// Вспомогательная функция для создания тестового пользователя
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
