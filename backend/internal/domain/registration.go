package domain

import (
	"time"
)

// registration entity
type Registration struct {
	ID           string     `json:"id"`
	EventID      string     `json:"event_id"`
	UserID       string     `json:"user_id"`
	Status       string     `json:"status"` // "registered", "cancelled", "waitlisted", "checked_in" also for the future btw
	RegisteredAt time.Time  `json:"registered_at"`
	CheckedInAt  *time.Time `json:"checked_in_at,omitempty"`
}

// DTOs
type RegisterForEventRequest struct {
	EventID string `json:"event_id" binding:"required,uuid"`
}

type CancelRegistrationRequest struct {
	EventID string `json:"event_id" binding:"required,uuid"`
}

// Constants for registration status
const (
	StatusRegistered = "registered"
	StatusCancelled  = "cancelled"
	StatusWaitlisted = "waitlisted"
	StatusCheckedIn  = "checked_in"
)
