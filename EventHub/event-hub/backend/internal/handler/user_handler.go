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

	protected.GET("/me", h.GetMe)
	protected.PATCH("/me", h.UpdateMe)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := c.GetString("userID") // из JWT

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	user, err := h.userService.UpdateMe(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, user)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("userID")

	user, err := h.userService.GetMe(userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, user)
}
