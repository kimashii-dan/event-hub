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

	// 4 & 5. Atomic Capacity Check and Creation
	// We pass the registration object (with UserID, EventID, Status) and the capacity limit.
	// The repository handles the locking and transaction.

	registration := &domain.Registration{
		UserID:  userID,
		EventID: eventID,
		Status:  "confirmed",
	}

	if err := s.regRepo.CreateWithCapacityCheck(registration, event.Capacity); err != nil {
		return nil, err // error is already formatted in repo
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

// CheckInAttendee marks user's attendance (organizer initiative)
func (s *RegistrationService) CheckInAttendee(organizerID, eventID, attendeeID string) error {

	// 1. Check if event exists and user is the organizer
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return fmt.Errorf("event not found")
	}

	if event.OrganizerID != organizerID {
		return fmt.Errorf("only the event organizer can check-in attendees")
	}

	// 2. Get the registration to verify it exists and is confirmed
	registration, err := s.regRepo.GetByUserAndEvent(attendeeID, eventID)
	if err != nil {
		return fmt.Errorf("failed to get registration: %w", err)
	}
	if registration == nil {
		return fmt.Errorf("registration not found")
	}

	// 3. Check if registration is in confirmed status
	if registration.Status != "confirmed" {
		return fmt.Errorf("can only check-in confirmed registrations (current status: %s)", registration.Status)
	}

	// 4. Update status to checked_in
	if err := s.regRepo.CheckIn(attendeeID, eventID); err != nil {
		return fmt.Errorf("failed to check-in attendee: %w", err)
	}

	return nil
}

// GetEventRegistrants returns all registrants for a specific event (organizer only)
func (s *RegistrationService) GetEventRegistrants(organizerID, eventID, status string) ([]domain.Registration, error) {

	// 1. Check if event exists and user is the organizer
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found")
	}

	if event.OrganizerID != organizerID {
		return nil, fmt.Errorf("only the event organizer can view registrants")
	}

	// 2. Get registrations with filtering
	registrations, err := s.regRepo.GetEventRegistrants(eventID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get event registrants: %w", err)
	}

	return registrations, nil
}
