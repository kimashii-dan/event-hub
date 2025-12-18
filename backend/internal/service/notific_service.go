package service

import (
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/Fixsbreaker/event-hub/backend/internal/worker"
	"github.com/google/uuid"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	pool             *worker.WorkerPool
}

func NewNotificationService(repo *repository.NotificationRepository, pool *worker.WorkerPool) *NotificationService {
	return &NotificationService{
		notificationRepo: repo,
		pool:             pool,
	}
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

	// Async send
	if s.pool != nil {
		s.pool.Submit(worker.NotificationJob{
			Notification: notification,
			DestEmail:    "user@example.com", // Stub for now
		})
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
