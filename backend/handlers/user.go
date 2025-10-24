package handlers

import (
	"net/http"
	"strconv"

	"github.com/booktracker/backend/middleware"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
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
			IsTeacher:     user.IsTeacher,
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
		IsTeacher:     user.IsTeacher,
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

	// Non-admin users cannot change admin or teacher status
	if currentUser != nil && !currentUser.IsAdmin {
		// Get current user data to preserve admin and teacher status
		existingUser, err := services.GetUserByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Message: err.Error(),
			})
			return
		}
		req.IsAdmin = existingUser.IsAdmin
		req.IsTeacher = existingUser.IsTeacher
	} else if currentUser != nil && currentUser.IsAdmin && currentUser.ID == uint(id) {
		// Prevent admins from disabling their own admin capability
		existingUser, err := services.GetUserByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Message: err.Error(),
			})
			return
		}
		if existingUser.IsAdmin && !req.IsAdmin {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Message: "You cannot remove your own admin privileges",
			})
			return
		}
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
		IsTeacher:     user.IsTeacher,
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

// MakeUserTeacher handles promoting a user to teacher role (admin only)
func MakeUserTeacher(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	user, err := services.MakeUserTeacher(uint(id))
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
		IsTeacher:     user.IsTeacher,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	c.JSON(http.StatusOK, userResponse)
}

// RemoveUserTeacher handles removing teacher role from a user (admin only)
func RemoveUserTeacher(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	user, err := services.RemoveUserTeacher(uint(id))
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
		IsTeacher:     user.IsTeacher,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	c.JSON(http.StatusOK, userResponse)
}

// CreateUser handles creating a new user (admin only)
func CreateUser(c *gin.Context) {
	var req models.CreateUserByAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	user, err := services.CreateUserByAdmin(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	// Send invitation email to set up password
	if !user.EmailVerified {
		// Get current admin user as inviter
		currentUser, _ := middleware.GetCurrentUser(c)
		if currentUser != nil {
			emailService := services.NewEmailService()
			err = emailService.SendSystemInvitationEmail(user.Email, currentUser.FirstName+" "+currentUser.LastName, user.EmailVerificationToken)
			if err != nil {
				// Don't fail user creation if email fails, but add a warning header
				c.Header("X-Email-Warning", "Invitation email failed to send")
			}
		}
	}

	userResponse := models.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		IsAdmin:       user.IsAdmin,
		IsTeacher:     user.IsTeacher,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	c.JSON(http.StatusCreated, userResponse)
}

// ResendUserVerificationEmail handles resending verification email for a specific user (admin only)
func ResendUserVerificationEmail(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid user ID",
		})
		return
	}

	// Get the user first to check if they need verification
	user, err := services.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	// Check if user already verified
	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "User email is already verified",
		})
		return
	}

	// Resend verification email
	updatedUser, err := services.ResendVerificationEmail(user.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	// Send the email
	emailService := services.NewEmailService()
	err = emailService.SendVerificationEmail(updatedUser.Email, updatedUser.FirstName, updatedUser.EmailVerificationToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to send verification email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent successfully",
	})
}