package domain

type CreateNotificationRequest struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type NotificationResponse struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Read    bool   `json:"read"`
}
