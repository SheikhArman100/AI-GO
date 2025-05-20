package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"my-project/internal/database"
	"my-project/internal/helper"
	"my-project/internal/model"
	"my-project/internal/oauth"
	"my-project/internal/response"
	"my-project/internal/validation"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db database.Service
}

func NewAuthHandler(db database.Service) *AuthHandler {
	return &AuthHandler{db: db}
}

// HelloAuth handles the GET request for auth root endpoint
func (h *AuthHandler) HelloAuth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello from auth group"})
}

// SignUp handles user registration
func (h *AuthHandler) SignUp(c *gin.Context) {
	// Get validated schema from context
	validated, exists := c.Get("validated")
	if !exists {
		response.ApiError(c, http.StatusInternalServerError, "Validation data missing")
		return
	}
	req, ok := validated.(*validation.SignUpRequest)
	if !ok {
		response.ApiError(c, http.StatusInternalServerError, "Invalid validation data")
		return
	}

	// Check for existing user
	var existingUser model.User
	if err := h.db.DB().Where("email = ? OR phone_number = ?", req.Email, req.PhoneNumber).First(&existingUser).Error; err == nil {
		response.ApiError(c, http.StatusUnprocessableEntity, "Email or PhoneNumber already exists")
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	// Create a new user
	user := &model.User{
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Password:    string(hashedPassword),
	}

	if err := h.db.DB().Create(user).Error; err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to registered new user", err.Error())
		return
	}

	// Generate email verification token
	expiresInStr := os.Getenv("JWT_EMAIL_VERIFY_EXPIRES_IN")
	expiresIn, err := strconv.ParseInt(expiresInStr, 10, 64)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Invalid JWT_EMAIL_VERIFY_EXPIRES_IN value", err.Error())
		return
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Second * time.Duration(expiresIn)).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_VERIFY_EMAIL_SECRET")))
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to generate verification token", err.Error())
		return
	}

	//Send verification email
	clientURL := os.Getenv("ADMIN_CLIENT_URL")
	emailBody := fmt.Sprintf(`
		<div>
			<p>Hi, %s</p>
			<p>Welcome to E-Commerce! Please verify your email address by clicking the link below:</p>
			<p>
				<a href="%s/auth/verify-email?token=%s">
					Verify Email
				</a>
			</p>
			<p>This link will expire soon.</p>
			<p>If you didn't create this account, you can ignore this email.</p>
			<p>Thank you, <br> E-Commerce</p>
		</div>`,
		user.Name, clientURL, token)

	if err := helper.SendEmail(user.Email, emailBody, "Verify Your Email"); err != nil {
		// response.ApiError(c, http.StatusInternalServerError, "Failed to send verification email", err.Error())
		// return
		log.Print("Failed to send verification email", err.Error())
	}

	// Send success response
	response.SendResponse(c, http.StatusCreated, true, "User registered successfully", gin.H{
		"user_id": user.ID,
	}, nil)
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")

	if token == "" {
		response.ApiError(c, http.StatusBadRequest, "Verification token is required")
		return
	}

	// Parse and validate the token
	claims := jwt.MapClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_VERIFY_EMAIL_SECRET")), nil
	})
	if err != nil || !t.Valid {

		response.ApiError(c, http.StatusBadRequest, "Invalid or expired verification token")
		return
	}

	// Extract user ID from the token
	idFloat, ok := claims["id"].(float64)
	if !ok {
		response.ApiError(c, http.StatusBadRequest, "Invalid token payload")
		return
	}
	userID := uint(idFloat)

	// Update user's email_verified_at field
	if err := h.db.DB().Model(&model.User{}).Where("id = ?", uint(userID)).Update("is_verified", true).Error; err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to verify email", err.Error())
		return
	}
	// Send success response
	response.SendResponse(c, http.StatusOK, true, "Email verified successfully", gin.H{
		"user_id": userID,
	}, nil)
}

// SignIn handles user authentication
func (h *AuthHandler) SignIn(c *gin.Context) {
	// Get validated schema from context
	validated, exists := c.Get("validated")
	if !exists {
		response.ApiError(c, http.StatusInternalServerError, "Validation data missing")
		return
	}
	req, ok := validated.(*validation.SignInRequest)
	if !ok {
		response.ApiError(c, http.StatusInternalServerError, "Invalid validation data")
		return
	}

	// Find user in database
	var user model.User
	if err := h.db.DB().Where("email =?", req.Email).First(&user).Error; err != nil {
		response.ApiError(c, http.StatusNotFound, "User doesn't exist.")
		return
	}
	//check user is verified or not
	if user.IsVerified == false {
		response.ApiError(c, http.StatusForbidden, "Please verify your email before signing in.")
		return
	}

	// Verify password using bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.ApiError(c, http.StatusUnauthorized, "Password is incorrect.")
		return
	}

	// Generate Access Token
	accessExpiresInStr := os.Getenv("JWT_ACCESS_TOKEN_EXPIRES_IN")
	accessExpiresIn, err := strconv.ParseInt(accessExpiresInStr, 10, 64)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Invalid JWT_ACCESS_TOKEN_EXPIRES_IN value", err.Error())
		return
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Second * time.Duration(accessExpiresIn)).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_ACCESS_TOKEN_SECRET")))
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to generate access token", err.Error())
		return
	}

	// Generate Refresh Token
	refreshExpiresInStr := os.Getenv("JWT_REFRESH_TOKEN_EXPIRES_IN")
	refreshExpiresIn, err := strconv.ParseInt(refreshExpiresInStr, 10, 64)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Invalid JWT_REFRESH_TOKEN_EXPIRES_IN value", err.Error())
		return
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Second * time.Duration(refreshExpiresIn)).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_REFRESH_TOKEN_SECRET")))
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to generate refresh token", err.Error())
		return
	}

	// Create a new refresh token record
	refreshTokenRecord := &model.RefreshToken{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Second * time.Duration(refreshExpiresIn)),
	}

	// Save refresh token to database
	if err := h.db.DB().Create(refreshTokenRecord).Error; err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to save refresh token", err.Error())
		return
	}

	// Set refresh token in cookies
	c.SetCookie("GO_JWT", refreshToken, int(refreshExpiresIn), "/", "", false, true)

	// Send success response with tokens
	response.SendResponse(c, http.StatusOK, true, "Sign in successful", gin.H{
		"access_token": accessToken,
	}, nil)
}

// update token
func (h *AuthHandler) UpdateToken(c *gin.Context) {
	// Get refresh token from cookies
	refreshToken, err := c.Cookie("GO_JWT")
	if err != nil {
		response.ApiError(c, http.StatusUnauthorized, "Please sign in first")
		return
	}

	// Verify refresh token
	claims := jwt.MapClaims{}
	t, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_REFRESH_TOKEN_SECRET")), nil

	})
	if err != nil || !t.Valid {
		response.ApiError(c, http.StatusBadRequest, "Invalid or expired refresh token")
		return
	}

	// Extract user ID from the token
	idFloat, ok := claims["id"].(float64)
	if !ok {
		response.ApiError(c, http.StatusBadRequest, "Invalid token payload")
		return
	}
	userID := uint(idFloat)

	// Check if the refresh token exists and belongs to the user
	var refreshTokenRecord model.RefreshToken
	if err := h.db.DB().Where("token = ? AND user_id = ?", refreshToken, userID).First(&refreshTokenRecord).Error; err != nil {
		response.ApiError(c, http.StatusBadRequest, "Invalid or expired refresh token")
		return
	}

	// Check if the refresh token is valid
	if refreshTokenRecord.ExpiresAt.Before(time.Now()) {
		response.ApiError(c, http.StatusBadRequest, "Invalid or expired refresh token")
		return
	}
	// Generate new access token
	accessExpiresInStr := os.Getenv("JWT_ACCESS_TOKEN_EXPIRES_IN")
	accessExpiresIn, err := strconv.ParseInt(accessExpiresInStr, 10, 64)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Invalid JWT_ACCESS_TOKEN_EXPIRES_IN value", err.Error())
		return
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userID,
		"email": claims["email"],
		"role":  claims["role"],
		"exp":   time.Now().Add(time.Second * time.Duration(accessExpiresIn)).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_ACCESS_TOKEN_SECRET")))
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to generate access token", err.Error())
		return
	}

	//generate new refresh token
	refreshExpiresInStr := os.Getenv("JWT_REFRESH_TOKEN_EXPIRES_IN")
	refreshExpiresIn, err := strconv.ParseInt(refreshExpiresInStr, 10, 64)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Invalid JWT_REFRESH_TOKEN_EXPIRES_IN value", err.Error())
		return
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userID,
		"email": claims["email"],
		"role":  claims["role"],
		"exp":   time.Now().Add(time.Second * time.Duration(refreshExpiresIn)).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_REFRESH_TOKEN_SECRET")))
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to generate refresh token", err.Error())
		return
	}
	// Update the refresh token in the database
	refreshTokenRecord.Token = refreshToken
	refreshTokenRecord.ExpiresAt = time.Now().Add(time.Second * time.Duration(refreshExpiresIn))
	if err := h.db.DB().Save(&refreshTokenRecord).Error; err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to update refresh token", err.Error())
		return
	}
	// Set refresh token in cookies
	c.SetCookie("GO_JWT", refreshToken, int(refreshExpiresIn), "/", "", false, true)

	// Send success response with tokens
	response.SendResponse(c, http.StatusOK, true, "Token updated successfully", gin.H{
		"access_token": accessToken,
	}, nil)

}

// signout
func (h *AuthHandler) SignOut(c *gin.Context) {
	// Get refresh token from cookies
	refreshToken, err := c.Cookie("GO_JWT")
	if err != nil {
		response.ApiError(c, http.StatusUnauthorized, "Please sign in first")
		return
	}

	// Verify refresh token
	claims := jwt.MapClaims{}
	t, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_REFRESH_TOKEN_SECRET")), nil

	})
	if err != nil || !t.Valid {
		response.ApiError(c, http.StatusBadRequest, "Invalid or expired refresh token")
		return
	}

	// Extract user ID from the token
	idFloat, ok := claims["id"].(float64)
	if !ok {
		response.ApiError(c, http.StatusBadRequest, "Invalid token payload")
		return
	}
	userID := uint(idFloat)

	// Delete the specific refresh token for this user from the database
	if err := h.db.DB().Where("token = ? AND user_id = ?", refreshToken, userID).Delete(&model.RefreshToken{}).Error; err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to sign out", err.Error())
		return
	}

	// Clear the refresh token cookie
	c.SetCookie("GO_JWT", "", -1, "/", "", false, true)
	fmt.Println("Sign out successful")

	// Send success response
	response.SendResponse(c, http.StatusOK, true, "Sign out successful", gin.H{
		"user_id": userID,
	}, nil)
}

// GoogleSignIn initiates the Google OAuth2 flow
func (h *AuthHandler) GoogleSignIn(c *gin.Context) {
	// Generate random state
	state := helper.GenerateRandomString(32)

	// Store state in cookie
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	// Get Google OAuth config
	googleConfig := oauth.GoogleOAuthConfig()

	// Redirect to Google's consent page
	url := googleConfig.AuthCodeURL(state)
	fmt.Println("Redirecting to Google OAuth2 consent page...", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the callback from Google
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Get state from cookie
	state, err := c.Cookie("oauth_state")
	if err != nil {
		response.ApiError(c, http.StatusBadRequest, "State cookie not found")
		return
	}

	// Verify state
	if state != c.Query("state") {
		response.ApiError(c, http.StatusBadRequest, "Invalid state parameter")
		return
	}

	// Exchange code for token
	code := c.Query("code")
	googleConfig := oauth.GoogleOAuthConfig()
	token, err := googleConfig.Exchange(c, code)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to exchange token", err.Error())
		return
	}

	// Get user info from Google
	googleUser, err := oauth.GetGoogleUser(token.AccessToken)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to get user info from Google", err.Error())
		return
	}

	// Start database transaction
	tx := h.db.DB().Begin()

	// Check if user exists
	var user model.User
	if err := tx.Where("email = ?", googleUser.Email).First(&user).Error; err != nil {
		// Generate random password for new user
		password := helper.GenerateRandomString(12)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			response.ApiError(c, http.StatusInternalServerError, "Failed to hash password", err.Error())
			return
		}

		// Create new user if not exists
		user = model.User{
			Name:       googleUser.Name,
			Email:      googleUser.Email,
			Password:   string(hashedPassword),
			IsVerified: true,
			Role:       "user",
		}
		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			response.ApiError(c, http.StatusInternalServerError, "Failed to create user", err.Error())
			return
		}

		// Send password via email
		emailBody := fmt.Sprintf(`
			<div>
				<p>Hi, %s</p>
				<p>Welcome to E-Commerce! Your account has been created with Google Sign-In.</p>
				<p>Your temporary password is: <strong>%s</strong></p>
				<p>Please change your password after signing in for security.</p>
				<p>Thank you, <br> E-Commerce</p>
			</div>`,
			user.Name, password)

		if err := helper.SendEmail(user.Email, emailBody, "Your E-Commerce Account Password"); err != nil {
			log.Print("Failed to send password email", err.Error())
		}
	}

	// Create or update social profile
	var socialProfile model.SocialProfile
	if err := tx.Where("user_id = ? AND provider = ?", user.ID, model.Google).First(&socialProfile).Error; err != nil {
		// Create new social profile
		socialProfile = model.SocialProfile{
			UserID:     user.ID,
			Provider:   model.Google,
			ProviderID: googleUser.ID,
			Name:       googleUser.Name,
			PhotoURL:   googleUser.Picture,
		}
		if err := tx.Create(&socialProfile).Error; err != nil {
			tx.Rollback()
			response.ApiError(c, http.StatusInternalServerError, "Failed to create social profile", err.Error())
			return
		}
	} else {
		// Update existing social profile
		socialProfile.Name = googleUser.Name
		socialProfile.PhotoURL = googleUser.Picture
		if err := tx.Save(&socialProfile).Error; err != nil {
			tx.Rollback()
			response.ApiError(c, http.StatusInternalServerError, "Failed to update social profile", err.Error())
			return
		}
	}

	// generate new refresh token
	refreshExpiresInStr := os.Getenv("JWT_REFRESH_TOKEN_EXPIRES_IN")
	refreshExpiresIn, err := strconv.ParseInt(refreshExpiresInStr, 10, 64)
	if err != nil {
		tx.Rollback()
		response.ApiError(c, http.StatusInternalServerError, "Invalid JWT_REFRESH_TOKEN_EXPIRES_IN value", err.Error())
		return
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Second * time.Duration(refreshExpiresIn)).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_REFRESH_TOKEN_SECRET")))
	if err != nil {
		tx.Rollback()
		response.ApiError(c, http.StatusInternalServerError, "Failed to generate refresh token", err.Error())
		return
	}

	// Create refresh token record
	refreshTokenRecord := model.RefreshToken{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Second * time.Duration(refreshExpiresIn)),
	}

	// Save refresh token to database
	if err := tx.Create(&refreshTokenRecord).Error; err != nil {
		tx.Rollback()
		response.ApiError(c, http.StatusInternalServerError, "Failed to save refresh token", err.Error())
		return
	}

	// Set refresh token in cookies
	c.SetCookie("GO_JWT", refreshToken, int(refreshExpiresIn), "/", "", false, true)
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	// Generate  access token
	accessExpiresInStr := os.Getenv("JWT_ACCESS_TOKEN_EXPIRES_IN")
	accessExpiresIn, err := strconv.ParseInt(accessExpiresInStr, 10, 64)
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Invalid JWT_ACCESS_TOKEN_EXPIRES_IN value", err.Error())
		return
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Second * time.Duration(accessExpiresIn)).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_ACCESS_TOKEN_SECRET")))
	if err != nil {
		response.ApiError(c, http.StatusInternalServerError, "Failed to generate access token", err.Error())
		return
	}

	// Send success response
	response.SendResponse(c, http.StatusOK, true, "Google sign in successful", gin.H{
		"access_token": accessToken,
		"user_id":      user.ID,
	}, nil)
}

// user details from refresh token
func (h *AuthHandler) UserDetails(c *gin.Context) {
	// Get refresh token from cookies
	refreshToken, err := c.Cookie("GO_JWT")
	if err != nil {
		response.ApiError(c, http.StatusUnauthorized, "Please sign in first")
		return
	}

	// Verify refresh token
	claims := jwt.MapClaims{}
	t, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_REFRESH_TOKEN_SECRET")), nil

	})
	if err != nil || !t.Valid {
		response.ApiError(c, http.StatusBadRequest, "Invalid or expired refresh token")
		return
	}

	//extract user id,role,email from claims
	userID := uint(claims["id"].(float64))
	role := claims["role"].(string)
	email := claims["email"].(string)

	//send success response
	response.SendResponse(c, http.StatusOK, true, "User details retrieved successfully", gin.H{
		"user_id": userID,
		"role":    role,
		"email":   email,
	}, nil)
}
