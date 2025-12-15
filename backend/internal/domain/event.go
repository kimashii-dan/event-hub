package domain

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Event entity with GORM tags
type Event struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrganizerID   string         `gorm:"type:uuid;not null;index" json:"organizer_id"`
	Organizer     *User          `gorm:"foreignKey:OrganizerID;constraint:OnDelete:CASCADE" json:"organizer,omitempty"` // Relationship
	Title         string         `gorm:"type:varchar(255);not null" json:"title"`
	Description   string         `gorm:"type:text" json:"description"`
	StartDatetime time.Time      `gorm:"not null;index" json:"start_datetime"`
	EndDatetime   time.Time      `gorm:"not null" json:"end_datetime"`
	Location      string         `gorm:"type:varchar(255);not null" json:"location"`
	Capacity      int            `gorm:"not null;check:capacity > 0" json:"capacity"`
	Status        string         `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"` // "draft", "published", "cancelled"
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (Event) TableName() string {
	return "events"
}

// DTOs
type CreateEventRequest struct {
	Title         string    `json:"title" binding:"required,min=3"`
	Description   string    `json:"description"`
	StartDatetime time.Time `json:"start_datetime" binding:"required"`
	EndDatetime   time.Time `json:"end_datetime" binding:"required"`
	Location      string    `json:"location" binding:"required"`
	Capacity      int       `json:"capacity" binding:"required,min=1"`
}

type UpdateEventRequest struct {
	Title         *string    `json:"title,omitempty"`
	Description   *string    `json:"description,omitempty"`
	StartDatetime *time.Time `json:"start_datetime,omitempty"`
	EndDatetime   *time.Time `json:"end_datetime,omitempty"`
	Location      *string    `json:"location,omitempty"`
	Capacity      *int       `json:"capacity,omitempty"`
}

// Validation
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

// Pagination structures
type PaginationRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type EventsResponse struct {
	Events     []Event             `json:"events"`
	Pagination *PaginationResponse `json:"pagination"`
}

type EventQueryRequest struct {
	// Pagination
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=20"`

	// Date filters
	StartDateFrom *time.Time `form:"start_date_from" time_format:"2006-01-02"`
	StartDateTo   *time.Time `form:"start_date_to" time_format:"2006-01-02"`

	// Capacity filters
	MinCapacity *int `form:"min_capacity" binding:"omitempty,min=1"`
	MaxCapacity *int `form:"max_capacity" binding:"omitempty,min=1"`

	// New filters
	Title        string `form:"title"`
	Status       string `form:"status"`
	Location     string `form:"location"`
	Keyword      string `form:"keyword"`
	OrganizerID  string `form:"organizer_id"`
	UpcomingOnly bool   `form:"upcoming_only"`
	PastOnly     bool   `form:"past_only"`

	// Sorting
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}
