package handlers

import (
	"net/http"

	"github.com/booktracker/backend-go/models"
	"github.com/booktracker/backend-go/services"
	"github.com/gin-gonic/gin"
)

// RegisterUser handles user registration
func RegisterUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	user, err := services.CreateUser(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	// Send verification email
	emailService := services.NewEmailService()
	err = emailService.SendVerificationEmail(user.Email, user.FirstName, user.EmailVerificationToken)
	if err != nil {
		// Don't fail registration if email fails, but log the error
		// In production you might want to queue this for retry
		c.Header("X-Email-Warning", "Verification email failed to send")
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

	c.JSON(http.StatusCreated, userResponse)
}

// LoginUser handles user login
func LoginUser(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	loginResponse, err := services.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "Invalid credentials",
		})
		return
	}

	c.JSON(http.StatusOK, loginResponse)
}