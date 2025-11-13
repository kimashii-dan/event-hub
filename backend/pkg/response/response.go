package response

import (
	"github.com/gin-gonic/gin"
)

// standard API Response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// success responses
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: true,
		Data: gin.H{
			"message": message,
		},
	})
}

// error responses
func Error(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// common error shortcuts
func BadRequest(c *gin.Context, message string) {
	Error(c, 400, "BAD_REQUEST", message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, "UNAUTHORIZED", message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, 403, "FORBIDDEN", message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, 404, "NOT_FOUND", message)
}

func Conflict(c *gin.Context, message string) {
	Error(c, 409, "CONFLICT", message)
}

func InternalServerError(c *gin.Context, message string) {
	Error(c, 500, "INTERNAL_SERVER_ERROR", message)
}
