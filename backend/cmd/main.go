//go:build server
// +build server

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/handlers"
	"github.com/booktracker/backend/middleware"
	"github.com/booktracker/backend/models"
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
				users.POST("", middleware.AdminMiddleware(), handlers.CreateUser)
				users.GET("", middleware.AdminMiddleware(), handlers.GetAllUsers)
				users.GET("/:id", handlers.GetUserByID)
				users.PUT("/:id", handlers.UpdateUser)
				users.DELETE("/:id", middleware.AdminMiddleware(), handlers.DeleteUser)
				users.PUT("/:id/make-teacher", middleware.AdminMiddleware(), handlers.MakeUserTeacher)
				users.PUT("/:id/remove-teacher", middleware.AdminMiddleware(), handlers.RemoveUserTeacher)
				users.POST("/:id/resend-verification", middleware.AdminMiddleware(), handlers.ResendUserVerificationEmail)
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
				
				// Book search route
				books.POST("/search", handlers.SearchBooks)
				
				// Create book from search result
				books.POST("/create-from-search", handlers.CreateBookFromSearch)
			}

			// Reports routes
			reports := protected.Group("/reports")
			{
				reports.GET("/my-books", handlers.GetMyBooksReport)
				reports.GET("/child/:childId/monthly-pdf", handlers.GenerateMonthlyPDFReport)
			}

			// Class routes
			classes := protected.Group("/classes")
			{
				classes.POST("", middleware.TeacherMiddleware(), handlers.CreateClass)
				classes.GET("", handlers.GetClasses)
				classes.GET("/available", handlers.GetAvailableClasses)
				classes.GET("/search-students", middleware.TeacherMiddleware(), handlers.SearchStudents)
				classes.GET("/:id", handlers.GetClass)
				classes.PUT("/:id", middleware.TeacherMiddleware(), handlers.UpdateClass)
				classes.DELETE("/:id", middleware.AdminMiddleware(), handlers.DeleteClass)
				classes.POST("/:id/members", middleware.TeacherMiddleware(), handlers.AddClassMember)
				classes.DELETE("/:id/members/:userId", middleware.TeacherMiddleware(), handlers.RemoveClassMember)
				classes.GET("/:id/students", handlers.GetClassStudents)
				classes.GET("/:id/teachers", handlers.GetClassTeachers)
				classes.POST("/assign-child", handlers.AssignChildToClass)
				classes.DELETE("/:id/children/:childId", middleware.TeacherMiddleware(), handlers.RemoveChildFromClass)
				
				// Student invitation routes
				classes.GET("/:id/invitation-data", middleware.TeacherMiddleware(), handlers.GetTeacherInvitationData)
				classes.POST("/:id/generate-invitation-token", middleware.TeacherMiddleware(), handlers.GenerateInvitationToken)
				classes.GET("/:id/google-sheets", middleware.TeacherMiddleware(), handlers.CreatePersonalizedGoogleSheet)
			}
			
			// Student invitation routes (some need to be public)
			invitations := protected.Group("/invitations")
			{
				invitations.GET("/pending", handlers.CheckPendingInvitation)
				invitations.POST("/redeem-pending", handlers.RedeemPendingInvitation)
			}
		}

		// Public student invitation routes (no auth required)
		api.GET("/invite/:classId/:token", handlers.GetStudentInvitationDetails)
		api.POST("/invite/:classId/:token/redeem", handlers.RedeemStudentInvitation)

		// Test routes setup (build tag controlled)
		setupTestRoutes(api) // Enable for e2e testing
	}

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(router.Run(":" + port))
}

// setupTestRoutes adds test-only routes for e2e testing
func setupTestRoutes(api *gin.RouterGroup) {
	test := api.Group("/test")
	{
		test.DELETE("/reset-db", func(c *gin.Context) {
			// Clear all tables for clean test state
			db := config.GetDB()
			
			// Delete in reverse dependency order
			db.Exec("DELETE FROM permissions")
			db.Exec("DELETE FROM pending_invitations") 
			db.Exec("DELETE FROM used_student_invitations")
			db.Exec("DELETE FROM books")
			db.Exec("DELETE FROM children")
			db.Exec("DELETE FROM class_memberships")
			db.Exec("DELETE FROM classes")
			db.Exec("DELETE FROM users")
			db.Exec("DELETE FROM shared_books")
			
			// Reset auto-increment counters
			db.Exec("DELETE FROM sqlite_sequence WHERE name IN ('users', 'children', 'books', 'classes', 'permissions', 'pending_invitations', 'used_student_invitations', 'class_memberships', 'shared_books')")
			
			c.JSON(200, gin.H{"message": "Database reset successful"})
		})
	}
}