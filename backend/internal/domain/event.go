package domain

import (
	"fmt"
	"time"
)

// event entity
type Event struct {
	ID            string    `json:"id"`
	OrganizerID   string    `json:"organizer_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	StartDatetime time.Time `json:"start_datetime"`
	EndDatetime   time.Time `json:"end_datetime"`
	Location      string    `json:"location"`
	Capacity      int       `json:"capacity"`
	Status        string    `json:"status"` // "draft", "published", "cancelled" but also it's more for the future
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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

// vvalidation
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
