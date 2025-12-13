package repository

import (
	"fmt"
	"time"

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

// GetEvents retrieves events with optional pagination and filters (generic & extensible)
func (r *EventRepository) GetEvents(req *domain.EventQueryRequest) ([]domain.Event, int64, error) {
	var events []domain.Event
	var total int64

	// Build base query
	query := r.db.Model(&domain.Event{})

	// Apply filters
	if req.StartDateFrom != nil {
		query = query.Where("start_datetime >= ?", *req.StartDateFrom)
	}
	if req.StartDateTo != nil {
		query = query.Where("start_datetime <= ?", *req.StartDateTo)
	}
	if req.MinCapacity != nil {
		query = query.Where("capacity >= ?", *req.MinCapacity)
	}
	if req.MaxCapacity != nil {
		query = query.Where("capacity <= ?", *req.MaxCapacity)
	}

	// --- NEW FILTERS ---
	if req.Title != "" {
		query = query.Where("title LIKE ?", "%"+req.Title+"%") // поиск по заголовку
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Location != "" {
		query = query.Where("location = ?", req.Location)
	}
	if req.OrganizerID != "" {
		query = query.Where("organizer_id = ?", req.OrganizerID)
	}
	if req.Keyword != "" {
		query = query.Where("description LIKE ?", "%"+req.Keyword+"%") // поиск по описанию
	}
	if req.UpcomingOnly {
		query = query.Where("start_datetime >= ?", time.Now())
	}
	if req.PastOnly {
		query = query.Where("end_datetime <= ?", time.Now())
	}

	// --- NEW SORTING ---
	sortColumn := "created_at"
	if req.SortBy != "" {
		sortColumn = req.SortBy
	}

	sortDirection := "DESC"
	if req.SortOrder == "asc" || req.SortOrder == "ASC" {
		sortDirection = "ASC"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortColumn, sortDirection))

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Apply pagination if requested
	if req.Page > 0 || req.PageSize > 0 {
		page := req.Page
		if page < 1 {
			page = 1
		}
		pageSize := req.PageSize
		if pageSize < 1 {
			pageSize = 10
		}
		if pageSize > 20 {
			pageSize = 20
		}
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// --- EXECUTE QUERY ---
	if err := query.Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch events: %w", err)
	}
	return events, total, nil
}
