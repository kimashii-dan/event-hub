package handler

import (
	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/Fixsbreaker/event-hub/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService *service.EventService
}

// NewEventHandler
// r               корневой роутер Gin
// eventService    сервис с бизнес-логикой и работой с репозиторием
// authMiddleware  middleware.Auth(config.JWTSecret)
func NewEventHandler(
	r *gin.Engine,
	eventService *service.EventService,
	authMiddleware gin.HandlerFunc,
) {
	h := &EventHandler{eventService: eventService}

	// Публичные роуты (чтобы смотреть события без авторизации)
	public := r.Group("/events")
	public.GET("", h.GetAllEvents)
	public.GET("/:id", h.GetEventByID)

	// Защищённые роуты (создание/редактирование/удаление только с токеном)
	protected := r.Group("/events")
	protected.Use(authMiddleware)
	protected.POST("", h.CreateEvent)
	protected.PUT("/:id", h.UpdateEvent)
	protected.DELETE("/:id", h.DeleteEvent)
	protected.POST("/:id/publish", h.PublishEvent)
	protected.POST("/:id/cancel", h.CancelEvent)
}

// helpers

// достаём user_id из контекста, который положил middleware.Auth
func getUserIDFromContext(c *gin.Context) (string, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	userID, ok := val.(string)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}

// handlers

// POST /events  (protected)
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req domain.CreateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unable to read user from context")
		return
	}

	event, err := h.eventService.CreateEvent(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, 201, event)
}

// PUT /events/:id  (protected)
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	var req domain.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unable to read user from context")
		return
	}

	event, err := h.eventService.UpdateEvent(userID, eventID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, 200, event)
}

// DELETE /events/:id  (protected)
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unable to read user from context")
		return
	}

	if err := h.eventService.DeleteEvent(userID, eventID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, 200, "event deleted")
}

// POST /events/:id/publish  (protected)
func (h *EventHandler) PublishEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unable to read user from context")
		return
	}

	if err := h.eventService.PublishEvent(userID, eventID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, 200, "event published")
}

// POST /events/:id/cancel  (protected)
func (h *EventHandler) CancelEvent(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unable to read user from context")
		return
	}

	if err := h.eventService.Cancel(userID, eventID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, 200, "event cancelled")
}

// GET /events  (public)
func (h *EventHandler) GetAllEvents(c *gin.Context) {
	events, err := h.eventService.GetAllEvents()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, 200, events)
}

// GET /events/:id  (public)
func (h *EventHandler) GetEventByID(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	event, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		// можно считать, что не нашли
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, 200, event)
}
