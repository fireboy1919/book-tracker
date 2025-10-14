package handlers

import (
	"net/http"
	"strconv"

	"github.com/booktracker/api/middleware"
	"github.com/booktracker/api/models"
	"github.com/booktracker/api/services"
	"github.com/gin-gonic/gin"
)

// GetAllUsers handles getting all users (admin only)
func GetAllUsers(c *gin.Context) {
	users, err := services.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to get users: " + err.Error(),
		})
		return
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, models.UserResponse{
			ID:            user.ID,
			Email:         user.Email,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			IsAdmin:       user.IsAdmin,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, userResponses)
}

// GetUserByID handles getting a user by ID
func GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	// Check if user is admin or requesting their own data
	currentUser, _ := middleware.GetCurrentUser(c)
	if currentUser != nil && !currentUser.IsAdmin && currentUser.ID != uint(id) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	user, err := services.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	userResponse := models.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		IsAdmin:       user.IsAdmin,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	c.JSON(http.StatusOK, userResponse)
}

// UpdateUser handles updating a user
func UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	// Check if user is admin or updating their own data
	currentUser, _ := middleware.GetCurrentUser(c)
	if currentUser != nil && !currentUser.IsAdmin && currentUser.ID != uint(id) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Non-admin users cannot change admin status
	if currentUser != nil && !currentUser.IsAdmin {
		// Get current user data to preserve admin status
		existingUser, err := services.GetUserByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Message: err.Error(),
			})
			return
		}
		req.IsAdmin = existingUser.IsAdmin
	}

	user, err := services.UpdateUser(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	userResponse := models.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		IsAdmin:       user.IsAdmin,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	c.JSON(http.StatusOK, userResponse)
}

// DeleteUser handles deleting a user
func DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	err = services.DeleteUser(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}