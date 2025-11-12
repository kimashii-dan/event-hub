package domain

import (
	"time"

	"gorm.io/gorm"
)

// Registration entity with GORM tags
type Registration struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	EventID      string         `gorm:"type:uuid;not null;index;uniqueIndex:idx_event_user" json:"event_id"`
	UserID       string         `gorm:"type:uuid;not null;index;uniqueIndex:idx_event_user" json:"user_id"`
	Event        *Event         `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE" json:"event,omitempty"`
	User         *User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Status       string         `gorm:"type:varchar(20);not null;default:'registered';index" json:"status"`
	RegisteredAt time.Time      `gorm:"autoCreateTime" json:"registered_at"`
	CheckedInAt  *time.Time     `gorm:"default:null" json:"checked_in_at,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (Registration) TableName() string {
	return "registrations"
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
