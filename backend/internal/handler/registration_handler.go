package handler

import (
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/Fixsbreaker/event-hub/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	regService *service.RegistrationService
}

func NewRegistrationHandler(r *gin.Engine, regService *service.RegistrationService, authMiddleware gin.HandlerFunc) {
	h := &RegistrationHandler{regService: regService}

	protected := r.Group("/")
	protected.Use(authMiddleware)

	// Register for event
	protected.POST("/events/:id/register", h.Register)

	// Cancel registration
	protected.DELETE("/events/:id/register", h.Cancel)

	// Get my registrations
	protected.GET("/users/me/registrations", h.GetMyRegistrations)

	// Check in attendee
	protected.PATCH("/events/:id/check/:attendee_id", h.CheckIn)

	// Get event registrants (organizer only), you can also filter with "?status=all/checked_in/confirmed/cancelled"
	protected.GET("/events/:id/registrants", h.GetEventRegistrants)
}

// POST /events/:id/register
func (h *RegistrationHandler) Register(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	reg, err := h.regService.RegisterUser(userID, eventID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, 201, reg)
}

// DELETE /events/:id/register
func (h *RegistrationHandler) Cancel(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.regService.CancelRegistration(userID, eventID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, 200, "registration cancelled")
}

// GET /users/me/registrations
func (h *RegistrationHandler) GetMyRegistrations(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	regs, err := h.regService.GetUserRegistrations(userID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, 200, regs)
}

// Mark attendee that he/she actually physically attends the event
func (h *RegistrationHandler) CheckIn(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	attendeeID := c.Param("attendee_id")
	if attendeeID == "" {
		response.BadRequest(c, "missing attendee id")
		return
	}

	organizerID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.regService.CheckInAttendee(organizerID, eventID, attendeeID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, 200, "attendee checked in successfully")
}

// GET /events/:id/registrants
func (h *RegistrationHandler) GetEventRegistrants(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		response.BadRequest(c, "missing event id")
		return
	}

	organizerID, ok := getUserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	// Get status filter from query parameter
	status := c.Query("status")
	if status == "" {
		status = "all"
	}

	// Validate status parameter
	validStatuses := map[string]bool{
		"all":        true,
		"confirmed":  true,
		"cancelled":  true,
		"checked_in": true,
	}
	if !validStatuses[status] {
		response.BadRequest(c, "invalid status filter. Valid values: all, confirmed, cancelled, checked_in")
		return
	}

	registrants, err := h.regService.GetEventRegistrants(organizerID, eventID, status)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, 200, registrants)
}
