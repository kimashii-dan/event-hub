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
