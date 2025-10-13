package main

import (
	"log"
	"net/http"
	"os"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/handlers"
	"github.com/booktracker/backend-go/middleware"
	"github.com/booktracker/backend-go/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	config.InitDatabase()
	
	// Auto-migrate the database
	err := models.AutoMigrate(config.GetDB())
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Setup Gin router
	router := gin.Default()

	// Setup CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))

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
			auth.POST("/login", handlers.LoginUser)
			auth.GET("/verify-email", handlers.VerifyEmail)
			auth.POST("/resend-verification", handlers.ResendVerification)
			auth.POST("/forgot-password", handlers.ForgotPassword)
			auth.POST("/reset-password", handlers.ResetPassword)
		}

		// Protected routes (authentication required)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
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
				books.GET("/child/:childId", handlers.GetBooksForChild)
			}

			// Reports routes
			reports := protected.Group("/reports")
			{
				reports.GET("/my-books", handlers.GetMyBooksReport)
			}
		}

		// Test routes setup (build tag controlled)
		setupTestRoutes(api)
	}

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(router.Run(":" + port))
}