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

// RegisterUserWithInvitation handles user registration via invitation
func RegisterUserWithInvitation(c *gin.Context) {
	var req models.CreateUserWithInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	user, err := services.ProcessInvitationRegistration(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	// Send verification email (optional, since they're invited)
	emailService := services.NewEmailService()
	err = emailService.SendVerificationEmail(user.Email, user.FirstName, user.EmailVerificationToken)
	if err != nil {
		// Don't fail registration if email fails
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

// GetInvitationDetails handles getting invitation details by token
func GetInvitationDetails(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invitation token is required",
		})
		return
	}

	invitation, err := services.GetPendingInvitationByToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	response := gin.H{
		"email":          invitation.Email,
		"childName":      invitation.Child.FirstName + " " + invitation.Child.LastName,
		"inviterName":    invitation.InvitedBy.FirstName + " " + invitation.InvitedBy.LastName,
		"permissionType": invitation.PermissionType,
	}

	c.JSON(http.StatusOK, response)
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

// ForgotPassword handles password reset request
func ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	user, err := services.RequestPasswordReset(req.Email)
	if err != nil {
		// Don't reveal if user exists or not for security
		c.JSON(http.StatusOK, gin.H{
			"message": "If an account with that email exists, a password reset email has been sent",
		})
		return
	}

	// Send password reset email
	emailService := services.NewEmailService()
	err = emailService.SendPasswordResetEmail(user.Email, user.FirstName, user.PasswordResetToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to send password reset email",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If an account with that email exists, a password reset email has been sent",
	})
}

// ResetPassword handles password reset with token
func ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	user, err := services.ResetPassword(req.Token, req.Password)
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
		"message": "Password reset successfully",
		"user":    userResponse,
	})
}