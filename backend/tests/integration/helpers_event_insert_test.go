package integration

import (
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// используем domain.Event напрямую
func insertEventDirectly(
	t *testing.T,
	db *gorm.DB,
	userID string,
) *domain.Event {
	t.Helper()

	event := &domain.Event{
		ID:            uuid.NewString(),
		OrganizerID:   userID,
		Title:         "Test Event",
		Description:   "Test Description",
		Location:      "Test Location",
		StartDatetime: time.Now().Add(time.Hour),
		EndDatetime:   time.Now().Add(2 * time.Hour),
		Capacity:      10,
		Status:        "draft",
	}

	// Напрямую вставляем domain.Event в БД
	if err := db.Create(event).Error; err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Проверяем что событие действительно создалось
	var check domain.Event
	if err := db.First(&check, "id = ?", event.ID).Error; err != nil {
		t.Fatalf("Event not found after insertion: %v", err)
	}

	return event
}
