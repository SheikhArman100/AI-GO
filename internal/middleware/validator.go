package middleware

import (
	"log"
	"net/http"

	"my-project/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidateRequest creates a middleware to validate the request body against a schema
func ValidateRequest(schema interface{}, validate *validator.Validate) gin.HandlerFunc {
	log.Print("ValidateRequest")
	return func(c *gin.Context) {
		// Bind JSON to schema
		if err := c.ShouldBind(schema); err != nil {
			response.ApiError(c, http.StatusBadRequest, "Invalid request payload", err.Error())
			return
		}

		// Validate schema
		if err := validate.Struct(schema); err != nil {
			response.ApiError(c, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}

		// Store validated schema in context for handler
		c.Set("validated", schema)
		c.Next()
	}
}
