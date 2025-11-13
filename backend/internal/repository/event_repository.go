package repository

import (
	"fmt"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

// create inserts a new event into the database
func (r *EventRepository) Create(event *domain.Event) error {
	result := r.db.Create(&event)
	if result.Error != nil {
		return fmt.Errorf("failed to create event: %w", result.Error)
	}
	return nil
}

// update modifies an existing event
func (r *EventRepository) Update(userID string, event *domain.Event) error {
	if event.OrganizerID != userID {
		return fmt.Errorf("user does not own the event")
	}
	result := r.db.Save(&event)
	if result.Error != nil {
		return fmt.Errorf("failed to update event: %w", result.Error)
	}
	return nil
}

// delete removes an event by ID
func (r *EventRepository) Delete(userID string, eventID string) error {
	result := r.db.Delete(&domain.Event{}, "id = ? AND organizer_id = ?", eventID, userID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete event: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found")
	}
	return nil
}

// publish event (set status to "published")
func (r *EventRepository) UpdateStatus(userID string, eventID string, newStatus string) error {
	result := r.db.Model(&domain.Event{}).Where("id = ? AND organizer_id = ?", eventID, userID).Update("status", newStatus)
	if result.Error != nil {
		return fmt.Errorf("failed to update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found or user does not own the event")
	}
	return nil
}

// getByID retrieves an event by ID
func (r *EventRepository) GetByID(id string) (*domain.Event, error) {
	var event domain.Event
	result := r.db.Where("id = ?", id).First(&event)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to get event by id: %w", result.Error)
	}
	return &event, nil
}

// getAll retrieves all events
func (r *EventRepository) GetAll() ([]domain.Event, error) {
	var events []domain.Event
	result := r.db.Find(&events)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get events: %w", result.Error)
	}
	return events, nil
}
