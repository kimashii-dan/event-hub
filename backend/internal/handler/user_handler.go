package handler

import (
	"net/http"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/Fixsbreaker/event-hub/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(r *gin.Engine, userService *service.UserService, authMiddleware gin.HandlerFunc) {
	h := &UserHandler{userService: userService}

	protected := r.Group("/users")
	protected.Use(authMiddleware)

	// PATCH /users/:id
	protected.PATCH("/:id", h.UpdateUser)
}

// UpdateUser handler
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "missing user id")
		return
	}

	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, user)
}
