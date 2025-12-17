package domain

import "time"

type Notification struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"` // к какому пользователю относится
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"` // прочитано или нет
	CreatedAt time.Time `json:"created_at"`
}
