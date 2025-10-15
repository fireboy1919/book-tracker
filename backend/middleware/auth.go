package middleware

import (
	"net/http"
	"strings"

	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := services.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Message: "Invalid token",
			})
			c.Abort()
			return
		}

		// Get user from database
		user, err := services.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Message: "User not found",
			})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("userId", user.ID)
		c.Next()
	}
}

// AdminMiddleware ensures user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Message: "User not found in context",
			})
			c.Abort()
			return
		}

		currentUser, ok := user.(*models.User)
		if !ok || !currentUser.IsAdmin {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Message: "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser helper function to get current user from context
func GetCurrentUser(c *gin.Context) (*models.User, error) {
	user, exists := c.Get("user")
	if !exists {
		return nil, nil
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		return nil, nil
	}

	return currentUser, nil
}

// GetCurrentUserID helper function to get current user ID from context
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userId")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}