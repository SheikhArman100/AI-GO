package server

import (
	"my-project/internal/handler"
	"my-project/internal/middleware"
	"my-project/internal/validation"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	// Global middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	// Add database middleware
	r.Use(middleware.DatabaseMiddleware(s.db))

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	//all routes for v1
	v1 := r.Group("/api/v1")
	{
		// Initialize handlers
		authHandler := handler.NewAuthHandler(s.db)
		userHandler := handler.NewUserHandler(s.db)

		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.GET("/", authHandler.HelloAuth)
			//sign up route
			auth.POST("/signup", middleware.ValidateRequest(&validation.SignUpRequest{}, validator.New()), authHandler.SignUp)
			//verify email route
			auth.PUT("/verify-email/:token", authHandler.VerifyEmail)
			//sign in route
			auth.POST("/signin", middleware.ValidateRequest(&validation.SignInRequest{}, validator.New()), authHandler.SignIn)
			//update token route
			auth.GET("/update-token", authHandler.UpdateToken)
			//sign out route
			auth.POST("/signout", authHandler.SignOut)
			//user details from token
			auth.GET("/user", authHandler.UserDetails)
			//Google OAuth routes
			auth.GET("/google/signin", authHandler.GoogleSignIn)
			auth.GET("/google/callback", authHandler.GoogleCallback)	
		}

		// User routes
		user := v1.Group("/user")
		{
			user.GET("/", userHandler.HelloUser)
			// Protected profile route
			user.GET("/profile", middleware.AuthMiddleware(), userHandler.GetProfile)
		}
	}

	//This route will catch the error if user hits a route that does not exist in our api.
	r.NoRoute(noRouteHandler)

	return r
}

// HelloWorldHandler handles the GET request for root endpoint
func (s *Server) HelloWorldHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
}

// healthHandler handles the GET request for health endpoint
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func noRouteHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"message": "Api not found!!! Wrong url, there is no route in this url."})
}
