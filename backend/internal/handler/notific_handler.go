package handler

import (
	"net/http"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/Fixsbreaker/event-hub/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(r *gin.Engine, service *service.NotificationService, authMiddleware gin.HandlerFunc) {
	h := &NotificationHandler{service: service}

	protected := r.Group("/notifications")
	protected.Use(authMiddleware)

	protected.GET("/", h.GetNotifications)
	protected.POST("/", h.SendNotification) // по желанию, можно оставить только админ
	protected.PATCH("/:id/read", h.MarkAsRead)
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetString("userID")

	notifications, err := h.service.GetUserNotifications(userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, notifications)
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	userID := c.GetString("userID") // или admin может указывать userID

	var req domain.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	notification, err := h.service.SendNotification(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, notification)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.MarkAsRead(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "notification marked as read"})
}
