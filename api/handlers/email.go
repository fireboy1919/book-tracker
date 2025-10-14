package handlers

import (
	"net/http"

	"github.com/booktracker/api/models"
	"github.com/booktracker/api/services"
	"github.com/gin-gonic/gin"
)

// VerifyEmail handles email verification
func VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Verification token is required",
		})
		return
	}

	user, err := services.VerifyEmail(token)
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

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
		"user":    userResponse,
	})
}

// ResendVerification handles resending verification emails
func ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	user, err := services.ResendVerificationEmail(req.Email)
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to send verification email: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent successfully",
	})
}