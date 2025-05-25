package helper

import (
    "errors"
    "github.com/gin-gonic/gin"
)

// GetValidatedFromContext retrieves type-safe validated data from Gin context
func GetValidatedFromContext[T any](c *gin.Context) (*T, error) {
    // Get validated data from context
    validated, exists := c.Get("validated")
    if !exists {
        return nil, errors.New("validation data missing")
    }

    // Type assert to generic type
    data, ok := validated.(*T)
    if !ok {
        return nil, errors.New("invalid validation data type")
    }

    return data, nil
}