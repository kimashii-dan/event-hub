package repository

import (
	"fmt"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create сохраняет новое уведомление
func (r *NotificationRepository) Create(notification *domain.Notification) error {
	result := r.db.Create(notification)
	if result.Error != nil {
		return fmt.Errorf("failed to create notification: %w", result.Error)
	}
	return nil
}

// GetByUserID возвращает все уведомления пользователя
func (r *NotificationRepository) GetByUserID(userID string) ([]domain.Notification, error) {
	var notifications []domain.Notification
	result := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", result.Error)
	}
	return notifications, nil
}

// MarkAsRead помечает уведомление как прочитанное
func (r *NotificationRepository) MarkAsRead(notificationID string) error {
	result := r.db.Model(&domain.Notification{}).Where("id = ?", notificationID).Update("read", true)
	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}
	return nil
}
