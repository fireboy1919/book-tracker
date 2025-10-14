package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/booktracker/backend-go/middleware"
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

	user, err := services.ProcessBulkInvitationRegistration(req)
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

// BulkInviteUser handles inviting a user to access multiple children with a single email
func BulkInviteUser(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	var req models.BulkInviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Verify that the current user owns all the children they're trying to share
	for _, childPerm := range req.Children {
		child, err := services.GetChildByID(childPerm.ChildID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Message: "Child not found",
			})
			return
		}

		if child.OwnerID != userID {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Message: "Only the owner can invite users to access their children",
			})
			return
		}
	}

	// Check if user already exists
	targetUser, err := services.GetUserByEmail(req.Email)
	if err != nil {
		// User doesn't exist - create pending invitations for all children
		token, err := services.CreateBulkPendingInvitation(req.Email, req.Children, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to create invitations: " + err.Error(),
			})
			return
		}

		// Send single invitation email with token
		currentUser, err := services.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to get current user: " + err.Error(),
			})
			return
		}

		err = services.SendSystemInvitationEmail(req.Email, token, currentUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to send invitation email: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Invitation sent successfully",
		})
		return
	}

	// User exists - create permissions directly
	for _, childPerm := range req.Children {
		err := services.CreateOrUpdatePermission(targetUser.ID, childPerm.ChildID, childPerm.PermissionType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to create permissions: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Access granted successfully",
	})
}

// Generate a random state string for OAuth
func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GoogleLogin handles the initial Google OAuth redirect
func GoogleLogin(c *gin.Context) {
	oauthService := services.NewOAuthService()
	
	state, err := generateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to generate state parameter",
		})
		return
	}

	// Store state in session/cookie for validation
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 minutes

	// Check if there's an invitation token
	invitationToken := c.Query("invitation_token")
	if invitationToken != "" {
		c.SetCookie("invitation_token", invitationToken, 600, "/", "", false, true)
	}

	authURL := oauthService.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback handles the OAuth callback from Google
func GoogleCallback(c *gin.Context) {
	oauthService := services.NewOAuthService()

	// Verify state parameter
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid state parameter",
		})
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Authorization code not provided",
		})
		return
	}

	// Exchange code for token
	token, err := oauthService.ExchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Failed to exchange code for token",
		})
		return
	}

	// Get user info from Google
	userInfo, err := oauthService.GetUserInfo(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Failed to get user information",
		})
		return
	}

	// Check if user already exists
	existingUser, err := services.GetUserByEmail(userInfo.Email)
	if err != nil {
		// User doesn't exist - check for invitation token
		invitationToken, err := c.Cookie("invitation_token")
		if err != nil {
			// No invitation token - create new user
			newUser, err := services.CreateGoogleUser(userInfo)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Message: "Failed to create user: " + err.Error(),
				})
				return
			}

			// Generate JWT token
			jwtToken, err := services.GenerateJWT(newUser.ID, newUser.Email)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Message: "Failed to generate token",
				})
				return
			}

			userResponse := models.UserResponse{
				ID:            newUser.ID,
				Email:         newUser.Email,
				FirstName:     newUser.FirstName,
				LastName:      newUser.LastName,
				IsAdmin:       newUser.IsAdmin,
				EmailVerified: newUser.EmailVerified,
				CreatedAt:     newUser.CreatedAt,
			}

			// Redirect to frontend with token and user info
			frontendURL := os.Getenv("FRONTEND_URL")
			if frontendURL == "" {
				frontendURL = "http://localhost:3000"
			}
			
			userJSON, _ := json.Marshal(userResponse)
			redirectURL := fmt.Sprintf("%s/google-callback?token=%s&user=%s", 
				frontendURL, jwtToken, url.QueryEscape(string(userJSON)))
			
			c.Redirect(http.StatusTemporaryRedirect, redirectURL)
			return
		}

		// Has invitation token - create user and process invitation
		c.SetCookie("invitation_token", "", -1, "/", "", false, true)
		
		newUser, err := services.CreateGoogleUserWithInvitation(userInfo, invitationToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Message: "Failed to create user with invitation: " + err.Error(),
			})
			return
		}

		// Generate JWT token
		jwtToken, err := services.GenerateJWT(newUser.ID, newUser.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to generate token",
			})
			return
		}

		userResponse := models.UserResponse{
			ID:            newUser.ID,
			Email:         newUser.Email,
			FirstName:     newUser.FirstName,
			LastName:      newUser.LastName,
			IsAdmin:       newUser.IsAdmin,
			EmailVerified: newUser.EmailVerified,
			CreatedAt:     newUser.CreatedAt,
		}

		// Redirect to frontend with token and user info
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}
		
		userJSON, _ := json.Marshal(userResponse)
		redirectURL := fmt.Sprintf("%s/google-callback?token=%s&user=%s", 
			frontendURL, jwtToken, url.QueryEscape(string(userJSON)))
		
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	// User exists - update OAuth fields if needed
	if existingUser.GoogleID == "" {
		err = services.LinkGoogleAccount(existingUser.ID, userInfo.ID, userInfo.Picture)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to link Google account: " + err.Error(),
			})
			return
		}
		existingUser.GoogleID = userInfo.ID
		existingUser.ProfilePicture = userInfo.Picture
	}

	// Generate JWT token
	jwtToken, err := services.GenerateJWT(existingUser.ID, existingUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to generate token",
		})
		return
	}

	userResponse := models.UserResponse{
		ID:            existingUser.ID,
		Email:         existingUser.Email,
		FirstName:     existingUser.FirstName,
		LastName:      existingUser.LastName,
		IsAdmin:       existingUser.IsAdmin,
		EmailVerified: existingUser.EmailVerified,
		CreatedAt:     existingUser.CreatedAt,
	}

	// Redirect to frontend with token and user info
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	
	userJSON, _ := json.Marshal(userResponse)
	redirectURL := fmt.Sprintf("%s/google-callback?token=%s&user=%s", 
		frontendURL, jwtToken, url.QueryEscape(string(userJSON)))
	
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}