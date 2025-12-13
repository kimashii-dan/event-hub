package integration

import (
	"testing"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/google/uuid"
)

// =====================
// UPDATE TESTS
// =====================

func TestUpdateEvent_Success(t *testing.T) {
	db, svc, cleanup := setupEventServiceTest(t)
	defer cleanup()

	userID := uuid.New().String()
	event := insertEventDirectly(t, db, userID)

	newTitle := "Updated Title"
	newDesc := "Updated Description"
	newCapacity := 25

	updated, err := svc.UpdateEvent(
		userID,   // ← сначала userID
		event.ID, // ← потом eventID
		&domain.UpdateEventRequest{
			Title:       &newTitle,
			Description: &newDesc,
			Capacity:    &newCapacity,
		},
	)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if updated.Title != newTitle {
		t.Errorf("Title not updated: expected '%s', got '%s'", newTitle, updated.Title)
	}
	if updated.Description != newDesc {
		t.Errorf("Description not updated: expected '%s', got '%s'", newDesc, updated.Description)
	}
	if updated.Capacity != newCapacity {
		t.Errorf("Capacity not updated: expected %d, got %d", newCapacity, updated.Capacity)
	}
}

func TestUpdateEvent_PartialUpdate(t *testing.T) {
	db, svc, cleanup := setupEventServiceTest(t)
	defer cleanup()

	userID := uuid.New().String()
	event := insertEventDirectly(t, db, userID)

	originalTitle := event.Title
	originalCapacity := event.Capacity

	newDesc := "Only description changed"

	updated, err := svc.UpdateEvent(
		userID,
		event.ID,
		&domain.UpdateEventRequest{
			Description: &newDesc,
		},
	)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if updated.Description != newDesc {
		t.Errorf("Description not updated: expected '%s', got '%s'", newDesc, updated.Description)
	}
	if updated.Title != originalTitle {
		t.Errorf("Title should not change: expected '%s', got '%s'", originalTitle, updated.Title)
	}
	if updated.Capacity != originalCapacity {
		t.Errorf("Capacity should not change: expected %d, got %d", originalCapacity, updated.Capacity)
	}
}

func TestUpdateEvent_NotOwner(t *testing.T) {
	db, svc, cleanup := setupEventServiceTest(t)
	defer cleanup()

	ownerID := uuid.New().String()
	otherUserID := uuid.New().String()

	event := insertEventDirectly(t, db, ownerID)

	hackedTitle := "Hacked Title"

	_, err := svc.UpdateEvent(
		otherUserID,
		event.ID,
		&domain.UpdateEventRequest{
			Title: &hackedTitle,
		},
	)

	if err == nil {
		t.Error("Expected error when non-owner tries to update event, got nil")
	}
}

func TestUpdateEvent_NonExistentEvent(t *testing.T) {
	_, svc, cleanup := setupEventServiceTest(t)
	defer cleanup()

	userID := uuid.New().String()
	fakeEventID := uuid.New().String()

	newTitle := "New Title"

	_, err := svc.UpdateEvent(
		userID,
		fakeEventID,
		&domain.UpdateEventRequest{
			Title: &newTitle,
		},
	)

	if err == nil {
		t.Error("Expected error when updating non-existent event, got nil")
	}
}

// =====================
// DELETE TESTS
// =====================

func TestDeleteEvent_Success(t *testing.T) {
	db, svc, cleanup := setupEventServiceTest(t)
	defer cleanup()

	userID := uuid.New().String()
	event := insertEventDirectly(t, db, userID)

	// ✅ ИСПРАВЛЕНО: правильный порядок (userID, eventID)
	err := svc.DeleteEvent(userID, event.ID)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	_, err = svc.GetEventByID(event.ID)
	if err == nil {
		t.Error("Expected error when getting deleted event, got nil")
	}
}

func TestDeleteEvent_NotOwner(t *testing.T) {
	db, svc, cleanup := setupEventServiceTest(t)
	defer cleanup()

	ownerID := uuid.New().String()
	otherUserID := uuid.New().String()

	event := insertEventDirectly(t, db, ownerID)

	err := svc.DeleteEvent(otherUserID, event.ID)
	if err == nil {
		t.Error("Expected error when non-owner tries to delete event, got nil")
	}
}

func TestDeleteEvent_NonExistentEvent(t *testing.T) {
	_, svc, cleanup := setupEventServiceTest(t)
	defer cleanup()

	userID := uuid.New().String()
	fakeEventID := uuid.New().String()

	err := svc.DeleteEvent(userID, fakeEventID)
	if err == nil {
		t.Error("Expected error when deleting non-existent event, got nil")
	}
}

func TestDeleteEvent_AlreadyDeleted(t *testing.T) {
	db, svc, cleanup := setupEventServiceTest(t)
	defer cleanup()

	userID := uuid.New().String()
	event := insertEventDirectly(t, db, userID)

	err := svc.DeleteEvent(userID, event.ID)
	if err != nil {
		t.Fatalf("First delete error: %v", err)
	}

	err = svc.DeleteEvent(userID, event.ID)
	if err == nil {
		t.Error("Expected error when deleting already deleted event, got nil")
	}
}
