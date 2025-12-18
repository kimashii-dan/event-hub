package service

import (
	"context"

	"fmt"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/cache"
	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
)

type EventService struct {
	eventRepo *repository.EventRepository
	cache     *cache.RedisCache
}

func NewEventService(eventRepo *repository.EventRepository, cache *cache.RedisCache) *EventService {
	return &EventService{
		eventRepo: eventRepo,
		cache:     cache,
	}
}

// event CRUD:
// create/update/delete events
// publish events
// get all events
// get event by id

// create event
func (s *EventService) CreateEvent(userID string, req *domain.CreateEventRequest) (*domain.Event, error) {
	event := &domain.Event{
		OrganizerID:   userID,
		Title:         req.Title,
		Description:   req.Description,
		Location:      req.Location,
		StartDatetime: req.StartDatetime,
		EndDatetime:   req.EndDatetime,
		Capacity:      req.Capacity,
		Status:        "draft",
	}

	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.eventRepo.Create(event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

// update event
func (s *EventService) UpdateEvent(userID string, eventID string, req *domain.UpdateEventRequest) (*domain.Event, error) {

	// fetch existing event
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// update fields if provided in request (only non-nil values)
	updated := false

	if req.Title != nil {
		event.Title = *req.Title
		updated = true
	}

	if req.Description != nil {
		event.Description = *req.Description
		updated = true
	}

	if req.StartDatetime != nil {
		event.StartDatetime = *req.StartDatetime
		updated = true
	}

	if req.EndDatetime != nil {
		event.EndDatetime = *req.EndDatetime
		updated = true
	}

	if req.Location != nil {
		event.Location = *req.Location
		updated = true
	}

	if req.Capacity != nil {
		event.Capacity = *req.Capacity
		updated = true
	}

	// If no fields were updated, return the existing event
	if !updated {
		return event, nil
	}

	// Validate the updated event
	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save the updated event
	if err := s.eventRepo.Update(userID, event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		if err := s.cache.Delete(context.Background(), fmt.Sprintf("event:%s", eventID)); err != nil {
			fmt.Printf("failed to invalidate cache for event %s: %v\n", eventID, err)
		}
	}

	return event, nil
}

// publish event
func (s *EventService) PublishEvent(userID string, eventID string) error {
	if err := s.eventRepo.UpdateStatus(userID, eventID, "published"); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	
	// Invalidate cache
	if err := s.cache.Delete(context.Background(), fmt.Sprintf("event:%s", eventID)); err != nil {
		fmt.Printf("failed to invalidate cache for event %s: %v\n", eventID, err)
	}

	return nil
}

// cancel event
func (s *EventService) Cancel(userID string, eventID string) error {
	if err := s.eventRepo.UpdateStatus(userID, eventID, "cancelled"); err != nil {
		return fmt.Errorf("failed to cancel event: %w", err)
	}
	return nil
}

// get event by id
func (s *EventService) GetEventByID(eventID string) (*domain.Event, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("event:%s", eventID)

	// 1. Try Cache
	var params domain.Event
	if s.cache != nil {
		if err := s.cache.Get(ctx, cacheKey, &params); err == nil {
			return &params, nil
		}
	}

	// 2. Fallback to DB
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// 3. Set Cache (ignore error)
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, event, 10*time.Minute); err != nil {
			fmt.Printf("failed to set cache for event %s: %v\n", eventID, err)
		}
	}

	return event, nil
}

// get all events
func (s *EventService) GetAllEvents() ([]domain.Event, error) {
	events, err := s.eventRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	return events, nil
}

// GetEvents returns events with optional pagination and filters
func (s *EventService) GetEvents(req *domain.EventQueryRequest) (*domain.EventsResponse, error) {
	// Set defaults
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 20 {
		req.PageSize = 20
	}
	if req.Page < 1 {
		req.Page = 1
	}

	// Basic validation
	if req.StartDateFrom != nil && req.StartDateTo != nil && req.StartDateTo.Before(*req.StartDateFrom) {
		return nil, fmt.Errorf("start_date_to must be after start_date_from")
	}
	if req.MinCapacity != nil && req.MaxCapacity != nil && *req.MaxCapacity < *req.MinCapacity {
		return nil, fmt.Errorf("max_capacity must be greater than or equal to min_capacity")
	}

	// Get events
	events, total, err := s.eventRepo.GetEvents(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// Build response
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	response := &domain.EventsResponse{
		Events: events,
		Pagination: &domain.PaginationResponse{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	return response, nil
}

// delete event
func (s *EventService) DeleteEvent(userID string, eventID string) error {
	if err := s.eventRepo.Delete(userID, eventID); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		if err := s.cache.Delete(context.Background(), fmt.Sprintf("event:%s", eventID)); err != nil {
			fmt.Printf("failed to invalidate cache for event %s: %v\n", eventID, err)
		}
	}

	return nil
}
