package integration

import (
	"time"

	"gorm.io/gorm"
)

// SQLITE-совместимая модель events
type eventTestModel struct {
	ID            string `gorm:"primaryKey"`
	OrganizerID   string `gorm:"index"`
	Title         string
	Description   string
	Location      string
	StartDatetime time.Time
	EndDatetime   time.Time
	Capacity      int
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (eventTestModel) TableName() string {
	return "events"
}

// SQLITE users
type userTestModel struct {
	ID        string `gorm:"primaryKey"`
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SQLITE registrations
type registrationTestModel struct {
	ID        string `gorm:"primaryKey"`
	UserID    string
	EventID   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}
