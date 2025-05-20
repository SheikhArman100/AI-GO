package middleware

import (
	"net/http"
	"os"
	"strings"

	"my-project/internal/database"
	"my-project/internal/model"
	"my-project/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware creates a middleware for protecting routes and optionally checking user roles
func AuthMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ApiError(c, http.StatusForbidden, "You are not authorized")
			c.Abort()
			return
		}

		// Check Bearer scheme
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.ApiError(c, http.StatusForbidden, "Invalid authorization format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_ACCESS_TOKEN_SECRET")), nil
		})

		if err != nil || !token.Valid {
			response.ApiError(c, http.StatusForbidden, "Invalid or expired token")
			c.Abort()
			return
		}

		// Get user ID from claims
		userID := uint(claims["id"].(float64))

		// Check if user is verified
		db := c.MustGet("db").(database.Service)
		var user model.User
		if err := db.DB().Select("is_verified").First(&user, userID).Error; err != nil {
			response.ApiError(c, http.StatusNotFound, "User not found")
			c.Abort()
			return
		}

		if !user.IsVerified {
			response.ApiError(c, http.StatusForbidden, "Please verify your email before accessing this resource")
			c.Abort()
			return
		}

		// Check required roles if any
		if len(requiredRoles) > 0 {
			userRole, ok := claims["role"].(string)
			if !ok {
				response.ApiError(c, http.StatusForbidden, "Invalid token payload")
				c.Abort()
				return
			}

			hasRole := false
			for _, role := range requiredRoles {
				if role == userRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				response.ApiError(c, http.StatusForbidden, "You have no access")
				c.Abort()
				return
			}
		}

		// Store user info in context
		c.Set("user", claims)
		c.Next()
	}
}

// DatabaseMiddleware injects the database service into the context
func DatabaseMiddleware(db database.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}
