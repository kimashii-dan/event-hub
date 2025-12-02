package domain

import (
	"time"

	"gorm.io/gorm"
)

// Registration entity
type Registration struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    string         `gorm:"type:uuid;not null;index" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	EventID   string         `gorm:"type:uuid;not null;index" json:"event_id"`
	Event     *Event         `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE" json:"event,omitempty"`
	Status    string         `gorm:"type:varchar(20);not null;default:'confirmed'" json:"status"` // "confirmed", "cancelled", "checked_in"
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for GORM
func (Registration) TableName() string {
	return "registrations"
}
