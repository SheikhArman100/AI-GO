package response

import (
	"runtime/debug"
	"os"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Success       bool        `json:"success"`
	StatusCode    int         `json:"statusCode"`
	Message       string      `json:"message"`
	ErrorMessages interface{} `json:"errorMessages,omitempty"`
	Stack         string      `json:"stack,omitempty"`
}

// ApiError sends an immediate error response and aborts further processing
func ApiError(c *gin.Context, statusCode int, message string, errorMessages ...interface{}) {
	stack := ""
	if os.Getenv("ENV") != "production" {
		stack = string(debug.Stack())
	}

	errResp := ErrorResponse{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
	}

	if len(errorMessages) > 0 {
		errResp.ErrorMessages = errorMessages[0]
	}

	if stack != "" {
		errResp.Stack = stack
	}

	c.AbortWithStatusJSON(statusCode, gin.H{"error": errResp})
}
