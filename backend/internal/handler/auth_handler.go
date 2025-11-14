package handler

import (
	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/Fixsbreaker/event-hub/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(r *gin.Engine, authService *service.AuthService) {
	h := &AuthHandler{authService: authService}
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
}

// Handler for register
func (h *AuthHandler) Register(c *gin.Context) {
	var body domain.CreateUserRequest

	// bind request body to Go struct
	if c.ShouldBindJSON(&body) != nil {
		response.BadRequest(c, "failed to read body")
		return
	}

	// use authService to register user
	user, err := h.authService.Register(&body)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// return created user
	response.Success(c, 201, user)
}

// Handler for login
func (h *AuthHandler) Login(c *gin.Context) {
	var body domain.LoginRequest

	// bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "failed to read body")
		return
	}

	// user authService to login user
	loginResponse, err := h.authService.Login(&body)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// return login response
	response.Success(c, 200, loginResponse)
}
