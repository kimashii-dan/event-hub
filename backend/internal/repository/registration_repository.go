package repository

import (
	"fmt"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"gorm.io/gorm"
)

type RegistrationRepository struct {
	db *gorm.DB
}

func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

// Create registers a user for an event
func (r *RegistrationRepository) Create(registration *domain.Registration) error {
	result := r.db.Create(registration)
	if result.Error != nil {
		return fmt.Errorf("failed to create registration: %w", result.Error)
	}
	return nil
}

// GetByUserAndEvent checks if a registration exists
func (r *RegistrationRepository) GetByUserAndEvent(userID, eventID string) (*domain.Registration, error) {
	var registration domain.Registration
	result := r.db.Where("user_id = ? AND event_id = ?", userID, eventID).First(&registration)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error here, just nil
		}
		return nil, fmt.Errorf("failed to get registration: %w", result.Error)
	}
	return &registration, nil
}

// GetUserRegistrations retrieves all registrations for a user
func (r *RegistrationRepository) GetUserRegistrations(userID string) ([]domain.Registration, error) {
	var registrations []domain.Registration
	// Preload Event data
	result := r.db.Preload("Event").Where("user_id = ?", userID).Find(&registrations)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user registrations: %w", result.Error)
	}
	return registrations, nil
}

// CancelRegistration cancels a registration (soft delete or status update)
// Here we will use soft delete as per GORM default, or update status if we want to keep history.
// Let's just delete the record for simplicity, or we can update status to 'cancelled'.
// Based on migration, we have a status field.
func (r *RegistrationRepository) Cancel(userID, eventID string) error {
	result := r.db.Model(&domain.Registration{}).
		Where("user_id = ? AND event_id = ?", userID, eventID).
		Update("status", "cancelled")

	if result.Error != nil {
		return fmt.Errorf("failed to cancel registration: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("registration not found")
	}
	return nil
}

// CountByEvent counts confirmed registrations for an event
func (r *RegistrationRepository) CountByEvent(eventID string) (int64, error) {
	var count int64
	result := r.db.Model(&domain.Registration{}).
		Where("event_id = ? AND status = ?", eventID, "confirmed").
		Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count registrations: %w", result.Error)
	}
	return count, nil
}
