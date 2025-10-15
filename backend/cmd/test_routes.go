//go:build !production
// +build !production

package main

import (
	"net/http"

	"github.com/booktracker/backend/config"
	"github.com/gin-gonic/gin"
)

func setupTestRoutes(api *gin.RouterGroup) {
	// Test routes (for compatibility) - only in non-production builds
	test := api.Group("/test")
	{
		test.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Test endpoint working"})
		})
		
		// Database reset endpoint for tests - only available in development/test builds
		test.DELETE("/reset-db", func(c *gin.Context) {
			db := config.GetDB()
			
			// Delete all data
			db.Exec("DELETE FROM permissions")
			db.Exec("DELETE FROM books")
			db.Exec("DELETE FROM children")
			db.Exec("DELETE FROM users")
			
			c.JSON(http.StatusOK, gin.H{"message": "Database reset successfully"})
		})
	}
}