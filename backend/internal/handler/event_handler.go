package handler

import (
	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/Fixsbreaker/event-hub/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// EventHandler handles HTTP requests for event-related operations.
// It serves as the presentation layer, delegating business logic to EventService.
type EventHandler struct {
	eventService *service.EventService
}

// NewEventHandler creates a new EventHandler and registers all event routes.
// It sets up both public and protected endpoints for event management.
//
// Parameters:
//   - r: The root Gin router to register routes on
//   - eventService: Service layer containing business logic for event operations
//   - authMiddleware: JWT authentication middleware for protected routes
//
// Public routes (no authentication required):
//   - GET /events - List all events with optional filtering and pagination
//   - GET /events/:id - Get a specific event by ID
//
// Protected routes (JWT authentication required):
//   - POST /events - Create a new event
//   - PUT /events/:id - Update an existing event (organizer only)
//   - DELETE /events/:id - Delete an event (organizer only)
//   - POST /events/:id/publish - Publish a draft event (organizer only)
//   - POST /events/:id/cancel - Cancel a published event (organizer only)
func NewEventHandler(
	r *gin.Engine,
	eventService *service.EventService,
	authMiddleware gin.HandlerFunc,
) {
	h := &EventHandler{eventService: eventService}

	// Public routes - accessible without authentication
	public := r.Group("/events")
	public.GET("", h.GetAllEvents)
	public.GET("/:id", h.GetEventByID)

	// Protected routes - require JWT authentication
	protected := r.Group("/events")
	protected.Use(authMiddleware)
	protected.POST("", h.CreateEvent)
	protected.PUT("/:id", h.UpdateEvent)
	protected.DELETE("/:id", h.DeleteEvent)
	protected.POST("/:id/publish", h.PublishEvent)
	protected.POST("/:id/cancel", h.CancelEvent)
}

// Helper Functions

// getUserIDFromContext extracts the user_id from the Gin context.
// The user_id is set by the authentication middleware after validating the JWT token.
//
// Returns:
//   - string: The user ID
//   - bool: True if user_id exists and is valid, false otherwise
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

// HTTP Handlers

// CreateEvent handles POST /events (protected)
// Creates a new event with the authenticated user as the organizer.
// The event is created in 'draft' status by default.
//
// Request Body: domain.CreateEventRequest
// Success Response: 201 Created with created event
// Error Responses:
//   - 400 Bad Request: Invalid input or validation failed
//   - 401 Unauthorized: Missing or invalid authentication
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req domain.CreateEventRequest

	// Parse and validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// Extract authenticated user ID
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unable to read user from context")
		return
	}

	// Create event through service layer
	event, err := h.eventService.CreateEvent(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, 201, event)
}

// UpdateEvent handles PUT /events/:id (protected)
// Updates an existing event. Only the event organizer can update their events.
//
// Path Parameters: id - Event UUID
// Request Body: domain.UpdateEventRequest (all fields optional)
// Success Response: 200 OK with updated event
// Error Responses:
//   - 400 Bad Request: Invalid input or event not found
//   - 401 Unauthorized: Missing authentication
//   - 403 Forbidden: User is not the event organizer
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

	// Service layer handles authorization check
	event, err := h.eventService.UpdateEvent(userID, eventID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, 200, event)
}

// DeleteEvent handles DELETE /events/:id (protected)
// Soft deletes an event. Only the event organizer can delete their events.
//
// Path Parameters: id - Event UUID
// Success Response: 200 OK with success message
// Error Responses:
//   - 400 Bad Request: Event not found
//   - 401 Unauthorized: Missing authentication
//   - 403 Forbidden: User is not the event organizer
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

// PublishEvent handles POST /events/:id/publish (protected)
// Changes event status from 'draft' to 'published', making it visible to public.
// Only the event organizer can publish their events.
//
// Path Parameters: id - Event UUID
// Success Response: 200 OK with success message
// Error Responses:
//   - 400 Bad Request: Event not found or already published
//   - 401 Unauthorized: Missing authentication
//   - 403 Forbidden: User is not the event organizer
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

// CancelEvent handles POST /events/:id/cancel (protected)
// Changes event status to 'cancelled'. Only the event organizer can cancel their events.
//
// Path Parameters: id - Event UUID
// Success Response: 200 OK with success message
// Error Responses:
//   - 400 Bad Request: Event not found or already cancelled
//   - 401 Unauthorized: Missing authentication
//   - 403 Forbidden: User is not the event organizer
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

// GetAllEvents handles GET /events (public)
// Retrieves a list of events with optional filtering, pagination, and sorting.
//
// Query Parameters:
//   - page: Page number (default: 1)
//   - page_size: Items per page (default: 10, max: 100)
//   - status: Filter by status (draft, published, cancelled)
//   - start_date_from: Filter events starting after this date (ISO 8601)
//   - start_date_to: Filter events starting before this date (ISO 8601)
//   - min_capacity: Minimum event capacity
//   - max_capacity: Maximum event capacity
//   - location: Filter by location (partial match)
//   - keyword: Search in title and description
//   - organizer_id: Filter by organizer UUID
//   - upcoming_only: Show only future events (true/false)
//   - past_only: Show only past events (true/false)
//   - sort_by: Sort field (start_date, capacity, created_at)
//
// Success Response: 200 OK
//   - With pagination: { "data": [...], "pagination": {...} }
//   - Without pagination: Array of events
//
// Error Responses:
//   - 400 Bad Request: Invalid query parameters
//   - 500 Internal Server Error: Database error
func (h *EventHandler) GetAllEvents(c *gin.Context) {
	var queryReq domain.EventQueryRequest

	// Parse and validate query parameters
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	// Check if any filtering/pagination parameters are provided
	hasParams := queryReq.Page > 0 || queryReq.PageSize > 0 ||
		queryReq.StartDateFrom != nil || queryReq.StartDateTo != nil ||
		queryReq.MinCapacity != nil || queryReq.MaxCapacity != nil ||
		queryReq.Status != "" || queryReq.Location != "" || queryReq.Keyword != "" ||
		queryReq.OrganizerID != "" || queryReq.UpcomingOnly || queryReq.PastOnly || queryReq.SortBy != ""

	// For backward compatibility, return all events if no parameters provided
	if !hasParams {
		events, err := h.eventService.GetAllEvents()
		if err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		response.Success(c, 200, events)
		return
	}

	// Apply filtering, pagination, and sorting
	eventsResponse, err := h.eventService.GetEvents(&queryReq)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, 200, eventsResponse)
}

// GetEventByID handles GET /events/:id (public)
// Retrieves a single event by its UUID. Includes organizer details.
//
// Path Parameters: id - Event UUID
// Success Response: 200 OK with event details
// Error Responses:
//   - 400 Bad Request: Missing event ID
//   - 404 Not Found: Event does not exist
func (h *EventHandler) GetEventByID(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	event, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, 200, event)
}

