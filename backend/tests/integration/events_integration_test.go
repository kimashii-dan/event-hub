package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/handler"
	"github.com/Fixsbreaker/event-hub/backend/internal/middleware"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ---------------------
// Test Setup with REAL repositories (in-memory SQLite)
// ---------------------
type testContext struct {
	router    *gin.Engine
	db        *gorm.DB
	eventRepo *repository.EventRepository
	regRepo   *repository.RegistrationRepository
	userRepo  *repository.UserRepository
}

func setupRouter(t *testing.T) *testContext {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Создаем in-memory SQLite БД
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Создаем таблицы
	setupTables(t, db)

	// Создаем РЕАЛЬНЫЕ репозитории (они работают с in-memory БД)
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)

	// Для Registration нужна таблица и репозиторий
	regRepo := setupRegistrationRepo(t, db)

	// Создаем сервисы с РЕАЛЬНЫМИ репозиториями
	authService := service.NewAuthService(userRepo, "test-secret", time.Hour)
	regService := service.NewRegistrationService(regRepo, eventRepo)

	// Регистрируем handlers
	handler.NewAuthHandler(r, authService)
	handler.NewRegistrationHandler(r, regService, middleware.Auth("test-secret"))

	return &testContext{
		router:    r,
		db:        db,
		eventRepo: eventRepo,
		regRepo:   regRepo,
		userRepo:  userRepo,
	}
}

func setupTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	// Users table
	usersSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        email TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        name TEXT NOT NULL,
        role TEXT NOT NULL DEFAULT 'user',
        created_at DATETIME,
        updated_at DATETIME,
        deleted_at DATETIME
    );`

	// Events table
	eventsSQL := `
    CREATE TABLE IF NOT EXISTS events (
        id TEXT PRIMARY KEY,
        organizer_id TEXT NOT NULL,
        title TEXT NOT NULL,
        description TEXT,
        start_datetime DATETIME NOT NULL,
        end_datetime DATETIME NOT NULL,
        location TEXT NOT NULL,
        capacity INTEGER NOT NULL,
        status TEXT NOT NULL DEFAULT 'draft',
        created_at DATETIME,
        updated_at DATETIME,
        deleted_at DATETIME
    );`

	// Registrations table
	registrationsSQL := `
    CREATE TABLE IF NOT EXISTS registrations (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        event_id TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'confirmed',
        registered_at DATETIME,
        created_at DATETIME,
        updated_at DATETIME,
        deleted_at DATETIME,
        UNIQUE(user_id, event_id)
    );`

	if err := db.Exec(usersSQL).Error; err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	if err := db.Exec(eventsSQL).Error; err != nil {
		t.Fatalf("Failed to create events table: %v", err)
	}
	if err := db.Exec(registrationsSQL).Error; err != nil {
		t.Fatalf("Failed to create registrations table: %v", err)
	}
}

func setupRegistrationRepo(t *testing.T, db *gorm.DB) *repository.RegistrationRepository {
	t.Helper()
	return repository.NewRegistrationRepository(db)
}

// ---------------------
// Helper Functions
// ---------------------

func createTestEvent(t *testing.T, eventRepo *repository.EventRepository, organizerID string) *domain.Event {
	t.Helper()

	event := &domain.Event{
		ID:            fmt.Sprintf("event-%d", time.Now().UnixNano()),
		OrganizerID:   organizerID,
		Title:         "Integration Test Event",
		Description:   "Event for integration testing",
		StartDatetime: time.Now().Add(24 * time.Hour),
		EndDatetime:   time.Now().Add(26 * time.Hour),
		Location:      "Test Location",
		Capacity:      100,
		Status:        "published",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := eventRepo.Create(event)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	return event
}

func registerUser(t *testing.T, router *gin.Engine, email, password, name string) {
	t.Helper()

	registerBody := map[string]string{
		"email":    email,
		"password": password,
		"name":     name,
	}
	bodyBytes, _ := json.Marshal(registerBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Registration failed: expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func loginUser(t *testing.T, router *gin.Engine, email, password string) string {
	t.Helper()

	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	bodyBytes, _ := json.Marshal(loginBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed: expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Parse response
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	// Extract "data"
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Login response missing 'data' field. Got: %v", resp)
	}

	// Extract token FROM data.token
	token, ok := data["token"].(string)
	if !ok || token == "" {
		t.Fatalf("Failed to extract token from login response. data = %v", data)
	}

	return token
}

func registerAndLogin(t *testing.T, router *gin.Engine, email, password, name string) string {
	t.Helper()
	registerUser(t, router, email, password, name)
	return loginUser(t, router, email, password)
}

// ---------------------
// Integration Tests
// ---------------------

func TestRegisterEvent_WithoutToken(t *testing.T) {
	ctx := setupRouter(t)
	event := createTestEvent(t, ctx.eventRepo, "organizer-123")

	req := httptest.NewRequest("POST", "/events/"+event.ID+"/register", nil)
	w := httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRegisterEvent_WithInvalidToken(t *testing.T) {
	ctx := setupRouter(t)
	event := createTestEvent(t, ctx.eventRepo, "organizer-123")

	req := httptest.NewRequest("POST", "/events/"+event.ID+"/register", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-xyz")
	w := httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for invalid token, got %d", w.Code)
	}
}

func TestRegisterEvent_Success(t *testing.T) {
	ctx := setupRouter(t)
	event := createTestEvent(t, ctx.eventRepo, "organizer-123")
	token := registerAndLogin(t, ctx.router, "user@example.com", "password123", "Test User")

	req := httptest.NewRequest("POST", "/events/"+event.ID+"/register", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	// Проверяем только HTTP статус
	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Если хочешь, можно ещё проверить, что тело не пустое
	if len(w.Body.Bytes()) == 0 {
		t.Error("Expected non-empty response body")
	}
}

func TestRegisterEvent_NonExistentEvent(t *testing.T) {
	ctx := setupRouter(t)
	token := registerAndLogin(t, ctx.router, "user2@example.com", "password123", "Test User 2")

	req := httptest.NewRequest("POST", "/events/non-existent-id/register", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 404 or 400 for non-existent event, got %d. Body: %s",
			w.Code, w.Body.String())
	}
}

func TestRegisterEvent_DuplicateRegistration(t *testing.T) {
	ctx := setupRouter(t)
	event := createTestEvent(t, ctx.eventRepo, "organizer-123")
	token := registerAndLogin(t, ctx.router, "user3@example.com", "password123", "Test User 3")

	// Первая регистрация
	req := httptest.NewRequest("POST", "/events/"+event.ID+"/register", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("First registration failed: expected 201, got %d. Body: %s",
			w.Code, w.Body.String())
	}

	// Вторая регистрация (дубликат)
	req = httptest.NewRequest("POST", "/events/"+event.ID+"/register", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 409 or 400 for duplicate, got %d. Body: %s",
			w.Code, w.Body.String())
	}
}

func TestRegisterEvent_FullCapacity(t *testing.T) {
	ctx := setupRouter(t)

	event := &domain.Event{
		ID:            fmt.Sprintf("limited-%d", time.Now().UnixNano()),
		OrganizerID:   "organizer-123",
		Title:         "Limited Event",
		Description:   "Only 1 spot",
		StartDatetime: time.Now().Add(24 * time.Hour),
		EndDatetime:   time.Now().Add(26 * time.Hour),
		Location:      "Test Location",
		Capacity:      1,
		Status:        "published",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	ctx.eventRepo.Create(event)

	token1 := registerAndLogin(t, ctx.router, "user4@example.com", "password123", "User 4")
	req := httptest.NewRequest("POST", "/events/"+event.ID+"/register", nil)
	req.Header.Set("Authorization", "Bearer "+token1)
	w := httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("First registration failed: got %d", w.Code)
	}

	token2 := registerAndLogin(t, ctx.router, "user5@example.com", "password123", "User 5")
	req = httptest.NewRequest("POST", "/events/"+event.ID+"/register", nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	w = httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusConflict {
		t.Errorf("Expected 400 or 409 for full capacity, got %d. Body: %s",
			w.Code, w.Body.String())
	}
}

func TestRegisterEvent_MultipleUsers(t *testing.T) {
	ctx := setupRouter(t)
	event := createTestEvent(t, ctx.eventRepo, "organizer-123")

	users := []struct {
		email    string
		password string
		name     string
	}{
		{"user6@example.com", "password123", "User 6"},
		{"user7@example.com", "password123", "User 7"},
		{"user8@example.com", "password123", "User 8"},
	}

	for _, user := range users {
		token := registerAndLogin(t, ctx.router, user.email, user.password, user.name)

		req := httptest.NewRequest("POST", "/events/"+event.ID+"/register", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		ctx.router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("User %s registration failed: expected 201, got %d",
				user.email, w.Code)
		}
	}
}
