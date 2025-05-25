package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// UserInfo represents the user information from JWT claims
type UserInfo struct {
    ID    uint   `json:"id"`
    Email string `json:"email"`
    Role  string `json:"role"`
}

// GetUserInfoFromContext extracts complete user info from JWT claims in Gin context
func GetUserInfoFromContext(c *gin.Context) (*UserInfo, error) {
    // Get user claims from context
    userData, exists := c.Get("user")
    if !exists {
        return nil, errors.New("user not authenticated")
    }

    // Convert to JWT claims
    claims, ok := userData.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid user data format")
    }

    // Extract user info fields
    userIDFloat, ok := claims["id"].(float64)
    if !ok {
        return nil, errors.New("invalid user ID format")
    }

    email, ok := claims["email"].(string)
    if !ok {
        return nil, errors.New("invalid email format")
    }

    role, ok := claims["role"].(string)
    if !ok {
        return nil, errors.New("invalid role format")
    }

    userInfo := &UserInfo{
        ID:    uint(userIDFloat),
        Email: email,
        Role:  role,
    }

    return userInfo, nil
}