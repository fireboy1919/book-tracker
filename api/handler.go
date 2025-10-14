//go:build serverless
// +build serverless

package main

import (
	"net/http"
	"sync"

	"github.com/booktracker/api/config"
	"github.com/booktracker/api/handlers"
	"github.com/booktracker/api/middleware"
	"github.com/booktracker/api/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
	once   sync.Once
)

func initRouter() {
	// Initialize database with optimized settings for serverless
	config.InitDatabase()
	
	// Auto-migrate the database
	err := models.AutoMigrate(config.GetDB())
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Setup Gin router
	router = gin.Default()

	// Setup CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))
	
	// Add permission cache middleware
	router.Use(middleware.PermissionCacheMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Health check endpoint for tests
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "OK"})
		})

		// Auth routes (no authentication required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterUser)
			auth.POST("/register-with-invitation", handlers.RegisterUserWithInvitation)
			auth.GET("/invitation-details", handlers.GetInvitationDetails)
			auth.POST("/login", handlers.LoginUser)
			auth.GET("/verify-email", handlers.VerifyEmail)
			auth.POST("/resend-verification", handlers.ResendVerification)
			auth.POST("/forgot-password", handlers.ForgotPassword)
			auth.POST("/reset-password", handlers.ResetPassword)
			
			// Google OAuth routes
			auth.GET("/google", handlers.GoogleLogin)
			auth.GET("/google/callback", handlers.GoogleCallback)
		}

		// Protected routes (authentication required)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Invitation routes
			protected.POST("/invite-user", handlers.BulkInviteUser)
			
			// User routes
			users := protected.Group("/users")
			{
				users.GET("", middleware.AdminMiddleware(), handlers.GetAllUsers)
				users.GET("/:id", handlers.GetUserByID)
				users.PUT("/:id", handlers.UpdateUser)
				users.DELETE("/:id", middleware.AdminMiddleware(), handlers.DeleteUser)
			}

			// Children routes
			children := protected.Group("/children")
			{
				children.POST("", handlers.CreateChild)
				children.GET("", handlers.GetChildren)
				children.GET("/with-counts", handlers.GetChildrenWithBookCounts)
				children.GET("/book-counts", handlers.GetBookCountsForChildren)
				children.GET("/:id", handlers.GetChildByID)
				children.PUT("/:id", handlers.UpdateChild)
				children.DELETE("/:id", handlers.DeleteChild)
				children.POST("/:id/invite", handlers.InviteUser)
				children.GET("/:id/permissions", handlers.GetPermissionsByChild)
			}

			// Permission routes
			permissions := protected.Group("/permissions")
			{
				permissions.DELETE("/:id", handlers.DeletePermissionByID)
			}

			// Books routes
			books := protected.Group("/books")
			{
				books.POST("", handlers.CreateBook)
				books.GET("", handlers.GetBooks)
				books.GET("/:id", handlers.GetBookByID)
				books.PUT("/:id", handlers.UpdateBook)
				books.DELETE("/:id", handlers.DeleteBook)
				
				// Child-specific book routes
				books.POST("/child/:childId", handlers.CreateBookForChild)
				books.POST("/child/:childId/custom", handlers.CreateCustomBookForChild)
				books.GET("/child/:childId", handlers.GetBooksForChild)
				
				// ISBN lookup route
				books.POST("/lookup-isbn", handlers.LookupISBN)
			}

			// Reports routes
			reports := protected.Group("/reports")
			{
				reports.GET("/my-books", handlers.GetMyBooksReport)
				reports.GET("/child/:childId/monthly-pdf", handlers.GenerateMonthlyPDFReport)
			}
		}
	}
}

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize router only once using sync.Once
	once.Do(initRouter)
	
	// Handle the request using Gin router
	router.ServeHTTP(w, r)
}