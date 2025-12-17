package domain

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Event represents an event entity in the system.
// Events can be created by users (organizers) and have different statuses:
//   - draft: Event is being created, not visible to public
//   - published: Event is live and accepting registrations
//   - cancelled: Event has been cancelled
//
// The entity uses GORM for ORM mapping and includes soft delete support.
type Event struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`              // Unique identifier (auto-generated UUID)
	OrganizerID   string         `gorm:"type:uuid;not null;index" json:"organizer_id"`                           // Foreign key to users table
	Organizer     *User          `gorm:"foreignKey:OrganizerID;constraint:OnDelete:CASCADE" json:"organizer,omitempty"` // Organizer user details (eager loaded when needed)
	Title         string         `gorm:"type:varchar(255);not null" json:"title"`                                // Event title (min 3 chars)
	Description   string         `gorm:"type:text" json:"description"`                                           // Detailed event description (optional)
	StartDatetime time.Time      `gorm:"not null;index" json:"start_datetime"`                                   // Event start date and time (indexed for date range queries)
	EndDatetime   time.Time      `gorm:"not null" json:"end_datetime"`                                           // Event end date and time
	Location      string         `gorm:"type:varchar(255);not null" json:"location"`                             // Event venue or location
	Capacity      int            `gorm:"not null;check:capacity > 0" json:"capacity"`                            // Maximum number of attendees (must be > 0)
	Status        string         `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`          // Event status: draft, published, cancelled (indexed for filtering)
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`                                       // Timestamp when event was created
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`                                       // Timestamp of last update
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`                                                         // Soft delete timestamp (null if not deleted)
}

// TableName specifies the database table name for the Event entity.
// GORM uses this to map the struct to the correct table.
func (Event) TableName() string {
	return "events"
}

// Data Transfer Objects (DTOs)

// CreateEventRequest represents the input data for creating a new event.
// All fields are validated using Gin's binding tags.
type CreateEventRequest struct {
	Title         string    `json:"title" binding:"required,min=3"`          // Event title (min 3 characters)
	Description   string    `json:"description"`                             // Event description (optional)
	StartDatetime time.Time `json:"start_datetime" binding:"required"`       // Event start date and time
	EndDatetime   time.Time `json:"end_datetime" binding:"required"`         // Event end date and time
	Location      string    `json:"location" binding:"required"`             // Event location/venue
	Capacity      int       `json:"capacity" binding:"required,min=1"`       // Maximum attendees (min 1)
}

// UpdateEventRequest represents the input data for updating an existing event.
// All fields are optional pointers - only non-nil fields will be updated.
type UpdateEventRequest struct {
	Title         *string    `json:"title,omitempty"`          // Update event title (optional)
	Description   *string    `json:"description,omitempty"`    // Update description (optional)
	StartDatetime *time.Time `json:"start_datetime,omitempty"` // Update start datetime (optional)
	EndDatetime   *time.Time `json:"end_datetime,omitempty"`   // Update end datetime (optional)
	Location      *string    `json:"location,omitempty"`       // Update location (optional)
	Capacity      *int       `json:"capacity,omitempty"`       // Update capacity (optional)
}

// Business Logic

// Validate performs business rule validation on the Event entity.
// This includes:
//   - Title must not be empty
//   - Capacity must be at least 1
//   - End time must be after start time
//   - Status defaults to "draft" if not set
//
// Returns an error if any validation rule fails.
func (e *Event) Validate() error {
	if e.Title == "" {
		return fmt.Errorf("title is required")
	}
	if e.Capacity < 1 {
		return fmt.Errorf("capacity must be at least 1")
	}
	if e.EndDatetime.Before(e.StartDatetime) {
		return fmt.Errorf("end time must be after start time")
	}
	if e.Status == "" {
		e.Status = "draft"
	}
	return nil
}

// Pagination Structures

// PaginationRequest contains pagination parameters for list queries.
type PaginationRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`              // Page number (starts at 1)
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"` // Items per page (max 100)
}

// PaginationResponse contains metadata about paginated results.
type PaginationResponse struct {
	Page       int   `json:"page"`        // Current page number
	PageSize   int   `json:"page_size"`   // Number of items per page
	Total      int64 `json:"total"`       // Total number of items across all pages
	TotalPages int   `json:"total_pages"` // Total number of pages
}

// EventsResponse wraps a list of events with pagination metadata.
type EventsResponse struct {
	Events     []Event             `json:"events"`     // List of event entities
	Pagination *PaginationResponse `json:"pagination"` // Pagination information
}

// EventQueryRequest contains all possible query parameters for filtering, sorting, and paginating events.
// All fields are optional and can be combined for complex queries.
type EventQueryRequest struct {
	// Pagination
	Page     int `form:"page" binding:"omitempty,min=1"`              // Page number (default: 1)
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=20"`  // Items per page (default: 10, max: 20)

	// Date filters
	StartDateFrom *time.Time `form:"start_date_from" time_format:"2006-01-02"` // Filter events starting after this date
	StartDateTo   *time.Time `form:"start_date_to" time_format:"2006-01-02"`   // Filter events starting before this date

	// Capacity filters
	MinCapacity *int `form:"min_capacity" binding:"omitempty,min=1"` // Minimum event capacity
	MaxCapacity *int `form:"max_capacity" binding:"omitempty,min=1"` // Maximum event capacity

	// Text filters
	Title       string `form:"title"`        // Filter by title (exact match)
	Status      string `form:"status"`       // Filter by status: draft, published, cancelled
	Location    string `form:"location"`     // Filter by location (partial match)
	Keyword     string `form:"keyword"`      // Search in title and description (partial match)
	OrganizerID string `form:"organizer_id"` // Filter by organizer user ID

	// Time-based filters
	UpcomingOnly bool `form:"upcoming_only"` // Show only future events (start_datetime > now)
	PastOnly     bool `form:"past_only"`     // Show only past events (end_datetime < now)

	// Sorting
	SortBy    string `form:"sort_by"`    // Field to sort by: start_date, capacity, created_at (default: created_at)
	SortOrder string `form:"sort_order"` // Sort direction: asc, desc (default: desc)
}

