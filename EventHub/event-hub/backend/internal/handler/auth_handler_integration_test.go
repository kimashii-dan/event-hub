package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Create isolated sqlite DB per test
func setupAuthDB(t *testing.T) *gorm.DB {
	dbFile := filepath.Join(t.TempDir(), "auth_test.db")

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
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
	if err := db.Exec(createTableSQL).Error; err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func setupAuthRouter(t *testing.T) *gin.Engine {
	db := setupAuthDB(t)
	repo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(repo, "test-secret", 1*time.Hour)

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	NewAuthHandler(r, authSvc)

	return r
}

func TestAuth_RegisterAndLogin(t *testing.T) {
	router := setupAuthRouter(t)

	// ----- Register -----
	registerBody := domain.CreateUserRequest{
		Email:    "test@example.com",
		Password: "mypassword123",
		Name:     "Test User",
	}

	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// ----- Login -----
	loginBody := domain.LoginRequest{
		Email:    "test@example.com",
		Password: "mypassword123",
	}

	b2, _ := json.Marshal(loginBody)
	req2 := httptest.NewRequest("POST", "/login", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")

	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", rec2.Code, rec2.Body.String())
	}

	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	if err := json.Unmarshal(rec2.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("invalid login JSON: %v", err)
	}

	if loginResp.Data.Token == "" {
		t.Fatalf("expected token, got empty: %s", rec2.Body.String())
	}
}

func TestAuth_InvalidJSON(t *testing.T) {
	router := setupAuthRouter(t)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte("{bad json")))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuth_DuplicateEmail(t *testing.T) {
	router := setupAuthRouter(t)

	// first registration
	user := domain.CreateUserRequest{
		Email:    "dup@example.com",
		Password: "pass123456",
		Name:     "User1",
	}
	b, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("unexpected: %d", rec.Code)
	}

	// second registration (same email)
	req2 := httptest.NewRequest("POST", "/register", bytes.NewReader(b))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 duplicate email, got %d", rec2.Code)
	}
}

func TestAuth_WrongPassword(t *testing.T) {
	router := setupAuthRouter(t)

	// create user
	user := domain.CreateUserRequest{
		Email:    "wrongpass@example.com",
		Password: "correct123",
		Name:     "Test",
	}
	b, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", rec.Code)
	}

	// login wrong password
	login := domain.LoginRequest{
		Email:    "wrongpass@example.com",
		Password: "incorrect",
	}
	b2, _ := json.Marshal(login)

	req2 := httptest.NewRequest("POST", "/login", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusUnauthorized && rec2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 or 401, got %d", rec2.Code)
	}
}
