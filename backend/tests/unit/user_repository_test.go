package unit

import (
	"database/sql"
	"testing"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
)

package unit

import (
"database/sql"
"testing"

"github.com/Fixsbreaker/event-hub/backend/internal/repository"
_ "github.com/mattn/go-sqlite3"
)

func TestUserRepository_Create(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	db.Exec(`CREATE TABLE users (
        id TEXT,
        email TEXT,
        password_hash TEXT,
        name TEXT,
        role TEXT,
        created_at DATETIME,
        updated_at DATETIME
    );`)

	repo := repository.NewUserRepository(db)

	u := &domain.User{
		ID:           "1",
		Email:        "email@test.com",
		PasswordHash: "hash",
		Name:         "John",
		Role:         "user",
	}

	err := repo.Create(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
