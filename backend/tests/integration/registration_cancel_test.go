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
type cancelTestContext struct {
	router    *gin.Engine
	db        *gorm.DB
	eventRepo *repository.EventRepository
	regRepo   *repository.RegistrationRepository
	userRepo  *repository.UserRepository
}

func setupCancelTestRouter(t *testing.T) *cancelTestContext {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Create tables
	setupTables(t, db)

	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	regRepo := repository.NewRegistrationRepository(db)

	authService := service.NewAuthService(userRepo, "test-secret", time.Hour)
	regService := service.NewRegistrationService(regRepo, eventRepo)

	handler.NewAuthHandler(r, authService)
	handler.NewRegistrationHandler(r, regService, middleware.Auth("test-secret"))

	return &cancelTestContext{
		router:    r,
		db:        db,
		eventRepo: eventRepo,
		regRepo:   regRepo,
		userRepo:  userRepo,
	}
}

// ---------------------
// Helper functions
// ---------------------

func createTestEventForCancel(t *testing.T, eventRepo *repository.EventRepository, organizerID string) *domain.Event {
	t.Helper()
	event := &domain.Event{
		ID:            fmt.Sprintf("event-%d", time.Now().UnixNano()),
		OrganizerID:   organizerID,
		Title:         "Cancel Test Event",
		Description:   "Event for testing cancel registration",
		StartDatetime: time.Now().Add(24 * time.Hour),
		EndDatetime:   time.Now().Add(26 * time.Hour),
		Location:      "Test Location",
		Capacity:      100,
		Status:        "published",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := eventRepo.Create(event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	return event
}

func registerAndLoginForCancel(t *testing.T, router *gin.Engine, email, password, name string) string {
	t.Helper()
	// Register user
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
		t.Fatalf("Registration failed: got %d", w.Code)
	}

	// Login user
	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	bodyBytes, _ = json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Login failed: got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Login response missing 'data' field")
	}

	token, ok := data["token"].(string)
	if !ok || token == "" {
		t.Fatalf("Failed to extract token from login response")
	}

	return token
}

// ---------------------
// Integration Tests
// ---------------------

func TestCancelRegistration_Success(t *testing.T) {
	ctx := setupCancelTestRouter(t)
	event := createTestEventForCancel(t, ctx.eventRepo, "organizer-123")
	token := registerAndLoginForCancel(t, ctx.router, "user-cancel@example.com", "password123", "Cancel User")

	// Register user to the event first
	req := httptest.NewRequest("POST", "/events/"+event.ID+"/register", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Initial registration failed: got %d", w.Code)
	}

	// Now cancel the registration
	req = httptest.NewRequest("DELETE", "/events/"+event.ID+"/register", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for cancel, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestCancelRegistration_WithoutToken(t *testing.T) {
	ctx := setupCancelTestRouter(t)
	event := createTestEventForCancel(t, ctx.eventRepo, "organizer-123")

	req := httptest.NewRequest("DELETE", "/events/"+event.ID+"/register", nil)
	w := httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized when token is missing, got %d", w.Code)
	}
}

func TestCancelRegistration_NotRegistered(t *testing.T) {
	ctx := setupCancelTestRouter(t)
	event := createTestEventForCancel(t, ctx.eventRepo, "organizer-123")
	token := registerAndLoginForCancel(t, ctx.router, "user-notreg@example.com", "password123", "Not Registered User")

	req := httptest.NewRequest("DELETE", "/events/"+event.ID+"/register", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ctx.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("Expected 400 or 404 when user is not registered, got %d", w.Code)
	}
}
