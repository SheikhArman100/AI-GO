package handler

import (
	"my-project/internal/database"
	"my-project/internal/model"
	"my-project/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type UserHandler struct {
	db database.Service
}

func NewUserHandler(db database.Service) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) HelloUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello from user group"})
}

// GetProfile handles fetching the authenticated user's profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user claims from context (set by auth middleware)
	userData, exists := c.Get("user")
	if !exists {
		response.ApiError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	claims, ok := userData.(jwt.MapClaims)
	if !ok {
		response.ApiError(c, http.StatusInternalServerError, "Invalid user data format")
		return
	}

	// Extract user ID from claims
	userID := uint(claims["id"].(float64))

	// Fetch user with related data
	var user model.User
	if err := h.db.DB().Preload("UserDetail").Preload("UserDetail.Image").Preload("SocialProfiles", func(db *gorm.DB) *gorm.DB {
		return db.Omit("User") // Skip loading User in SocialProfiles since it's redundant (user data is already in the parent object)
	}).First(&user, userID).Error; err != nil {
		response.ApiError(c, http.StatusNotFound, "User not found")
		return
	}

	// Send success response
	response.SendResponse(c, http.StatusOK, true, "Profile retrieved successfully", user, nil)
}
