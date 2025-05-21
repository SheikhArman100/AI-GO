package handler

import (
	"errors"
	"fmt"
	"my-project/internal/database"
	"my-project/internal/helper"
	"my-project/internal/model"
	"my-project/internal/response"
	"my-project/internal/validation"
	"net/http"
	"path/filepath"

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
	if err := h.db.DB().Preload("UserDetail").Preload("UserDetail.Image").Preload("SocialProfiles").First(&user, userID).Error; err != nil {
		response.ApiError(c, http.StatusNotFound, "User not found")
		return
	}

	// Send success response
	response.SendResponse(c, http.StatusOK, true, "Profile retrieved successfully", user, nil)
}

// UpdateProfile handles updating the authenticated user's profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	//get userinfo from context (set by auth middleware)
	userInterface, exist := c.Get("user")
	if !exist {
		response.ApiError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	// Type assertion to jwt.MapClaims
	userInfo, ok := userInterface.(jwt.MapClaims)
	if !ok {
		response.ApiError(c, http.StatusInternalServerError, "Invalid user data format")
		return
	}

	// Safely extract user ID
	idFloat, ok := userInfo["id"].(float64)
	if !ok {
		response.ApiError(c, http.StatusInternalServerError, "User ID not found or invalid")
		return
	}

	userID := uint(idFloat) // Convert float64 to uint

	//get request body from context (set by validation middleware)
	validatedRequest, exist := c.Get("validated")
	if !exist {
		response.ApiError(c, http.StatusInternalServerError, "Invalid request data")
		return
	}

	// Type assertion to UpdateProfileRequest
	req, ok := validatedRequest.(*validation.UpdateProfileRequest)
	if !ok {
		response.ApiError(c, http.StatusInternalServerError, "Invalid request data format")
		return
	}

	///transaction started
	db := h.db.DB()
	tx := db.Begin()
	if tx.Error != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to start transaction: "+tx.Error.Error())
		return
	}
	//fetch user with related data
	var user model.User
	if err := tx.Preload("UserDetail").Preload("UserDetail.Image").First(&user, userID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ApiError(c, http.StatusNotFound, "User not found")
		} else {
			response.ApiError(c, http.StatusInternalServerError, "Failed to fetch user: "+err.Error())
		}
		return
	}
	fmt.Printf("Request body:\n%+v\n", req)

	// Update User model fields
	userUpdated := false
	if req.Name != "" && user.Name != req.Name {
		user.Name = req.Name
		userUpdated = true
	}

	if userUpdated {
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			response.ApiError(c, http.StatusInternalServerError, "Failed to update user name: "+err.Error())
			return
		}
	}

	// Ensure UserDetail exists or create it
	if user.UserDetail == nil {
		user.UserDetail = &model.UserDetail{UserID: user.ID}
		// This new UserDetail will be saved below if any fields are set or if it's forced by an image upload later
	}

	userDetailUpdated := false
	if req.Address != "" && user.UserDetail.Address != req.Address {
		user.UserDetail.Address = req.Address
		userDetailUpdated = true
	}
	if req.City != "" && user.UserDetail.City != req.City {
		user.UserDetail.City = req.City
		userDetailUpdated = true
	}
	if req.Road != "" && user.UserDetail.Road != req.Road {
		user.UserDetail.Road = req.Road
		userDetailUpdated = true
	}

	// Handle Image Upload
	if req.Image != nil {
		uploadDir := filepath.Join(".", "upload", "user") // Relative to project root
		//check file type is image ore not
		if !helper.IsImageFile(req.Image) {
			tx.Rollback()
			response.ApiError(c, http.StatusBadRequest, "Invalid file type. Only image files are allowed.")
			return
		}
		// Upload image locally
		uploadedFile, err := helper.UploadFileLocally(req.Image, uploadDir)
		if err != nil {
			tx.Rollback()
			response.ApiError(c, http.StatusInternalServerError, "Failed to upload image: "+err.Error())
			return
		}

		if user.UserDetail.Image != nil {
			// Update existing image record
			user.UserDetail.Image.Path = uploadedFile.WebPath
			user.UserDetail.Image.DiskType = model.DiskType(uploadedFile.DiskType)
			user.UserDetail.Image.OriginalName = uploadedFile.OriginalName
			user.UserDetail.Image.ModifiedName = uploadedFile.ModifiedName
			fmt.Println("Image record updated:", user.UserDetail.Image)
			if err := tx.Save(user.UserDetail.Image).Error; err != nil {
				tx.Rollback()
				response.ApiError(c, http.StatusInternalServerError, "Failed to update image record: "+err.Error())
				return
			}
		} else {
			// Create new image record
			newImage := model.Image{
				UserDetailID: &user.UserDetail.ID,
				Path:         uploadedFile.WebPath,
				DiskType:     model.DiskType(uploadedFile.DiskType),
				OriginalName: uploadedFile.OriginalName,
				ModifiedName: uploadedFile.ModifiedName,
			}
			if err := tx.Create(&newImage).Error; err != nil {
				tx.Rollback()
				response.ApiError(c, http.StatusInternalServerError, "Failed to create image record: "+err.Error())
				return
			}
			user.UserDetail.Image = &newImage
			fmt.Println("Image record created:", user.UserDetail.Image)
		}
		userDetailUpdated = true
	}

	//* Save UserDetail if it's new (ID=0) or if fields were updated (including image)
	if user.UserDetail != nil && (user.UserDetail.ID == 0 || userDetailUpdated) {
		if err := tx.Save(user.UserDetail).Error; err != nil {
			tx.Rollback()
			response.ApiError(c, http.StatusInternalServerError, "Failed to update user details: "+err.Error())
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	// Refetch user with all updated details for the response
	var updatedUser model.User
	if err := db.Preload("UserDetail").Preload("UserDetail.Image").Preload("SocialProfiles").First(&updatedUser, userID).Error; err != nil {
		// Log this error but still return a success message as the update itself was successful if commit was ok
		response.ApiError(c, http.StatusInternalServerError, "Failed to fetch updated user: "+err.Error())
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Profile updated successfully", updatedUser, nil)
}
