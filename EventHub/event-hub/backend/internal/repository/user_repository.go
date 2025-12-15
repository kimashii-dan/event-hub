package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// create inserts a new user into the database
func (r *UserRepository) Create(user *domain.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	return nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", result.Error)
	}
	return &user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by id: %w", result.Error)
	}
	return &user, nil
}

// EmailExists checks if an email already exists in the database
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	result := r.db.Model(&domain.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check email existence: %w", result.Error)
	}
	return count > 0, nil
}

// UpdateRole updates a user's role (useful for admin operations)
func (r *UserRepository) UpdateRole(userID, newRole string) error {
	result := r.db.Model(&domain.User{}).Where("id = ?", userID).Update("role", newRole)
	if result.Error != nil {
		return fmt.Errorf("failed to update user role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// Delete soft-deletes a user (GORM automatically uses DeletedAt)
func (r *UserRepository) Delete(userID string) error {
	result := r.db.Delete(&domain.User{}, "id = ?", userID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// GetAll retrieves all users (admin operation, можно добавить пагинацию)
func (r *UserRepository) GetAll(limit, offset int) ([]domain.User, error) {
	var users []domain.User
	result := r.db.Limit(limit).Offset(offset).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get users: %w", result.Error)
	}
	return users, nil
}

// Update updates user information
func (r *UserRepository) Update(user *domain.User) error {
	result := r.db.Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	return nil
}
