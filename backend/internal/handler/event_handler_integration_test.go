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
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// create per-test sqlite file DB to avoid shared in-memory DB between tests
func setupInMemoryDB(t *testing.T) *gorm.DB {
	dbFile := filepath.Join(t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	require.NoError(t, err)

	createSQL := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		organizer_id TEXT,
		title TEXT,
		description TEXT,
		location TEXT,
		start_datetime DATETIME,
		end_datetime DATETIME,
		capacity INTEGER,
		status TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	);`
	err = db.Exec(createSQL).Error
	require.NoError(t, err)

	return db
}

func TestCreateAndGetEvents_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupInMemoryDB(t)
	repo := repository.NewEventRepository(db)
	svc := service.NewEventService(repo)

	h := &EventHandler{eventService: svc}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	req := domain.CreateEventRequest{
		Title:         "Integration Test Event",
		Description:   "desc",
		Location:      "loc",
		StartDatetime: time.Now().Add(time.Hour),
		EndDatetime:   time.Now().Add(2 * time.Hour),
		Capacity:      10,
	}
	body, err := json.Marshal(req)
	require.NoError(t, err)

	c.Request = httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "tester1")

	h.CreateEvent(c)
	require.Equal(t, http.StatusCreated, rec.Code)

	events, err := svc.GetAllEvents()
	require.NoError(t, err)
	require.Len(t, events, 1)
}

func TestGetAllEvents_Integration_EmptyAndMultiple(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupInMemoryDB(t)
	repo := repository.NewEventRepository(db)
	svc := service.NewEventService(repo)
	h := &EventHandler{eventService: svc}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/events", nil)

	h.GetAllEvents(c)
	require.Equal(t, http.StatusOK, rec.Code)

	_, err := svc.CreateEvent("u1", &domain.CreateEventRequest{Title: "E1", StartDatetime: time.Now(), EndDatetime: time.Now().Add(time.Hour), Capacity: 1})
	require.NoError(t, err)
	_, err = svc.CreateEvent("u1", &domain.CreateEventRequest{Title: "E2", StartDatetime: time.Now(), EndDatetime: time.Now().Add(time.Hour), Capacity: 2})
	require.NoError(t, err)

	rec2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(rec2)
	c2.Request = httptest.NewRequest(http.MethodGet, "/events", nil)

	h.GetAllEvents(c2)
	require.Equal(t, http.StatusOK, rec2.Code)

	var resp2 map[string]interface{}
	err = json.Unmarshal(rec2.Body.Bytes(), &resp2)
	require.NoError(t, err)
	arr, ok := resp2["data"].([]interface{})
	require.True(t, ok)
	require.Len(t, arr, 2)
}
