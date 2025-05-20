package response

import (
	"github.com/gin-gonic/gin"
)

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Count int `json:"count"`
}

type IApiResponse[T any] struct {
	StatusCode int    `json:"statusCode"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	Meta       *Meta  `json:"meta,omitempty"`
	Data       T      `json:"data,omitempty"`
}

func SendResponse[T any](c *gin.Context, statusCode int, success bool, message string, data T, meta *Meta) {
	response := IApiResponse[T]{
		StatusCode: statusCode,
		Success:    success,
		Message:    message,
		Meta:       meta,
		Data:       data,
	}

	c.JSON(statusCode, response)
}
