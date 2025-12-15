package service

import (
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/google/uuid"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notificationRepo: repo}
}

// Отправка уведомления пользователю
func (s *NotificationService) SendNotification(userID string, req *domain.CreateNotificationRequest) (*domain.Notification, error) {
	notification := &domain.Notification{
		ID:        uuid.NewString(),
		UserID:    userID,
		Title:     req.Title,
		Message:   req.Message,
		Read:      false,
		CreatedAt: time.Now(),
	}
	if err := s.notificationRepo.Create(notification); err != nil {
		return nil, err
	}
	return notification, nil
}

// Получение всех уведомлений пользователя
func (s *NotificationService) GetUserNotifications(userID string) ([]domain.Notification, error) {
	return s.notificationRepo.GetByUserID(userID)
}

// Пометить уведомление как прочитанное
func (s *NotificationService) MarkAsRead(notificationID string) error {
	return s.notificationRepo.MarkAsRead(notificationID)
}
