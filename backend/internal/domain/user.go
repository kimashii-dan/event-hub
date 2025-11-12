package domain

import (
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

// User entity with GORM tags
type User struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"` // Never expose in JSON
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Role         string         `gorm:"type:varchar(20);not null;default:'user'" json:"role"` // "user", "organizer", "admin"
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete support
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// DTOs
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      *User     `json:"user"`
}

// Basic validation
func (u *User) Validate() error {
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !isValidEmail(u.Email) {
		return fmt.Errorf("invalid email format")
	}
	if u.Name == "" {
		return fmt.Errorf("name is required")
	}
	if u.Role == "" {
		u.Role = "user" // default
	}
	return nil
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
