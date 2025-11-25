package service

import (
	"fmt"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
)

type RegistrationService struct {
	regRepo   *repository.RegistrationRepository
	eventRepo *repository.EventRepository
}

func NewRegistrationService(regRepo *repository.RegistrationRepository, eventRepo *repository.EventRepository) *RegistrationService {
	return &RegistrationService{
		regRepo:   regRepo,
		eventRepo: eventRepo,
	}
}

// RegisterUser registers a user for an event
func (s *RegistrationService) RegisterUser(userID, eventID string) (*domain.Registration, error) {
	// 1. Check if event exists
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found")
	}

	// 2. Check if event is published
	if event.Status != "published" {
		return nil, fmt.Errorf("cannot register for unpublished event")
	}

	// 3. Check if already registered
	existing, err := s.regRepo.GetByUserAndEvent(userID, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing registration: %w", err)
	}
	if existing != nil {
		if existing.Status == "cancelled" {
			// If previously cancelled, we could re-activate, but for now let's just say "already registered"
			// or we could update the status back to confirmed.
			// Let's keep it simple: error.
			return nil, fmt.Errorf("user already registered (status: %s)", existing.Status)
		}
		return nil, fmt.Errorf("user already registered")
	}

	// 4. Check capacity
	count, err := s.regRepo.CountByEvent(eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to check capacity: %w", err)
	}

	if int(count) >= event.Capacity {
		return nil, fmt.Errorf("event is full")
	}

	// 5. Create registration
	registration := &domain.Registration{
		UserID:  userID,
		EventID: eventID,
		Status:  "confirmed",
	}

	if err := s.regRepo.Create(registration); err != nil {
		return nil, fmt.Errorf("failed to create registration: %w", err)
	}

	return registration, nil
}

// CancelRegistration cancels a user's registration
func (s *RegistrationService) CancelRegistration(userID, eventID string) error {
	return s.regRepo.Cancel(userID, eventID)
}

// GetUserRegistrations returns all events a user is registered for
func (s *RegistrationService) GetUserRegistrations(userID string) ([]domain.Registration, error) {
	return s.regRepo.GetUserRegistrations(userID)
}
