package response

import (
	"fmt"
	"my-project/internal/logger"
	"runtime"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	Success       bool        `json:"success"`
	StatusCode    int         `json:"statusCode"`
	Message       string      `json:"message"`
	ErrorMessages interface{} `json:"errorMessages,omitempty"`
	Stack         string      `json:"stack,omitempty"`
}

func GetCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// ApiError sends an immediate error response and aborts further processing
func ApiError(c *gin.Context, statusCode int, message string, errorMessages ...interface{}) {
	stack := GetCallerInfo(2)

	// Get user ID from context if available
	user, _ := c.Get("user")

	// Log error with context information
	logger.ErrorLogger.Error("API Error",
		zap.Int("statusCode", statusCode),
		zap.String("message", message),
		zap.String("ip", c.ClientIP()),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Any("user", user),
		zap.String("stack", stack),
	)

	errResp := ErrorResponse{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
	}

	if len(errorMessages) > 0 {
		errResp.ErrorMessages = errorMessages[0]
		// Log additional error messages
		logger.ErrorLogger.Error("Additional error details",
			zap.Any("errorMessages", errorMessages[0]))
	}

	if stack != "" {
		errResp.Stack = stack
	}

	c.AbortWithStatusJSON(statusCode, gin.H{"error": errResp})
}
