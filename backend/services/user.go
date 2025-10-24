package services

import (
	"errors"
	"strings"
	"time"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/utils"
	"gorm.io/gorm"
)

// CreateUser creates a new user
func CreateUser(req models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	result := config.GetDB().Where("email = ?", req.Email).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Generate verification token
	token, err := utils.GenerateVerificationToken()
	if err != nil {
		return nil, err
	}
	
	expiresAt := utils.GetTokenExpiration()

	// Determine admin status
	var isAdmin bool
	if req.IsAdmin {
		// Explicit admin request (used by tests)
		isAdmin = true
	} else {
		// Check if this is the first user (should be admin)
		var userCount int64
		err := config.GetDB().Model(&models.User{}).Count(&userCount).Error
		if err != nil {
			return nil, err
		}
		
		if userCount == 0 {
			// This is the first user - make them admin
			isAdmin = true
		} else {
			isAdmin = false
		}
	}

	// Create user
	user := models.User{
		Email:                  req.Email,
		PasswordHash:           passwordHash,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		IsAdmin:                isAdmin,
		EmailVerified:          false,       // New users need to verify their email
		EmailVerificationToken: token,
		TokenExpiresAt:         &expiresAt,
	}

	result = config.GetDB().Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByID gets a user by ID
func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := config.GetDB().First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByEmail gets a user by email
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := config.GetDB().Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetAllUsers gets all users (admin only)
func GetAllUsers() ([]models.User, error) {
	var users []models.User
	result := config.GetDB().Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// UpdateUser updates a user
func UpdateUser(id uint, req models.UpdateUserRequest) (*models.User, error) {
	var user models.User
	result := config.GetDB().First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	// Check if email is already taken by another user
	var existingUser models.User
	result = config.GetDB().Where("email = ? AND id != ?", req.Email, id).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("email already taken by another user")
	}

	// Update user
	user.Email = req.Email
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.IsAdmin = req.IsAdmin
	user.IsTeacher = req.IsTeacher

	result = config.GetDB().Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// DeleteUser deletes a user
func DeleteUser(id uint) error {
	result := config.GetDB().Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// VerifyEmail verifies a user's email address using the verification token
func VerifyEmail(token string) (*models.User, error) {
	var user models.User
	result := config.GetDB().Where("email_verification_token = ?", token).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Check if this might be a user who is already verified
			var verifiedUser models.User
			emailResult := config.GetDB().Where("email_verified = ? AND email_verification_token = ''", true).First(&verifiedUser)
			if emailResult.Error == nil {
				return nil, errors.New("email address is already verified")
			}
			return nil, errors.New("invalid verification token")
		}
		return nil, result.Error
	}

	// Check if already verified (shouldn't happen, but safety check)
	if user.EmailVerified {
		return nil, errors.New("email address is already verified")
	}

	// Check if token has expired
	if utils.IsTokenExpired(user.TokenExpiresAt) {
		return nil, errors.New("verification token has expired")
	}

	// Mark email as verified and clear token
	user.EmailVerified = true
	user.EmailVerificationToken = ""
	user.TokenExpiresAt = nil

	result = config.GetDB().Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// ResendVerificationEmail generates a new verification token for a user
func ResendVerificationEmail(email string) (*models.User, error) {
	var user models.User
	result := config.GetDB().Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	// Check if already verified
	if user.EmailVerified {
		return nil, errors.New("email is already verified")
	}

	// Generate new token
	token, err := utils.GenerateVerificationToken()
	if err != nil {
		return nil, err
	}
	
	expiresAt := utils.GetTokenExpiration()
	user.EmailVerificationToken = token
	user.TokenExpiresAt = &expiresAt

	result = config.GetDB().Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByVerificationToken gets a user by their verification token
func GetUserByVerificationToken(token string) (*models.User, error) {
	var user models.User
	result := config.GetDB().Where("email_verification_token = ?", token).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid verification token")
		}
		return nil, result.Error
	}
	return &user, nil
}

// RequestPasswordReset generates a password reset token for a user
func RequestPasswordReset(email string) (*models.User, error) {
	var user models.User
	result := config.GetDB().Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	// Generate reset token
	token, err := utils.GenerateVerificationToken()
	if err != nil {
		return nil, err
	}
	
	// Set expiration to 1 hour from now
	expiresAt := time.Now().Add(1 * time.Hour)
	user.PasswordResetToken = token
	user.PasswordResetExpiresAt = &expiresAt

	result = config.GetDB().Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// ResetPassword resets a user's password using the reset token
func ResetPassword(token, newPassword string) (*models.User, error) {
	var user models.User
	result := config.GetDB().Where("password_reset_token = ?", token).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid reset token")
		}
		return nil, result.Error
	}

	// Check if token has expired
	if user.PasswordResetExpiresAt == nil || time.Now().After(*user.PasswordResetExpiresAt) {
		return nil, errors.New("reset token has expired")
	}

	// Hash the new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return nil, err
	}

	// Update password and clear reset token
	user.PasswordHash = hashedPassword
	user.PasswordResetToken = ""
	user.PasswordResetExpiresAt = nil

	result = config.GetDB().Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// CreateGoogleUser creates a new user from Google OAuth info
func CreateGoogleUser(userInfo *GoogleUserInfo) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	result := config.GetDB().Where("email = ?", userInfo.Email).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Determine admin status
	var userCount int64
	err := config.GetDB().Model(&models.User{}).Count(&userCount).Error
	if err != nil {
		return nil, err
	}
	
	isAdmin := userCount == 0 // First user is admin

	// Create user
	user := models.User{
		Email:          userInfo.Email,
		FirstName:      userInfo.GivenName,
		LastName:       userInfo.FamilyName,
		IsAdmin:        isAdmin,
		EmailVerified:  userInfo.VerifiedEmail, // Trust Google's verification
		GoogleID:       userInfo.ID,
		AuthProvider:   "google",
		ProfilePicture: userInfo.Picture,
	}

	result = config.GetDB().Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// CreateGoogleUserWithInvitation creates a new user from Google OAuth and processes invitation
func CreateGoogleUserWithInvitation(userInfo *GoogleUserInfo, invitationToken string) (*models.User, error) {
	// Try to detect if this is a student invitation token first
	studentInvitationService := NewStudentInvitationService(config.GetDB())
	_, err := studentInvitationService.DecryptInvitationToken(invitationToken)
	
	if err == nil {
		// This is a student invitation - create user and redeem invitation
		return createGoogleUserWithStudentInvitation(userInfo, invitationToken, studentInvitationService)
	}

	// Not a student invitation - handle legacy invitation system
	invitations, err := GetPendingInvitationsByToken(invitationToken)
	if err != nil {
		return nil, err
	}

	if len(invitations) == 0 {
		return nil, errors.New("invalid invitation token")
	}

	// Verify email matches invitation
	if !strings.EqualFold(invitations[0].Email, userInfo.Email) {
		return nil, errors.New("email does not match invitation")
	}

	// Create user with Google OAuth info
	user := models.User{
		Email:          userInfo.Email,
		FirstName:      userInfo.GivenName,
		LastName:       userInfo.FamilyName,
		IsAdmin:        false, // Invited users are not admin by default
		EmailVerified:  userInfo.VerifiedEmail,
		GoogleID:       userInfo.ID,
		AuthProvider:   "google",
		ProfilePicture: userInfo.Picture,
	}

	// Start transaction
	tx := config.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create user
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Process invitations and create permissions
	for _, invitation := range invitations {
		permission := models.Permission{
			UserID:         user.ID,
			ChildID:        invitation.ChildID,
			PermissionType: invitation.PermissionType,
		}

		if err := tx.Create(&permission).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Delete processed invitations
	if err := tx.Where("token = ?", invitationToken).Delete(&models.PendingInvitation{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// createGoogleUserWithStudentInvitation handles Google OAuth registration with student invitation
func createGoogleUserWithStudentInvitation(userInfo *GoogleUserInfo, invitationToken string, studentInvitationService *StudentInvitationService) (*models.User, error) {
	// First, validate the invitation token and get the payload
	payload, err := studentInvitationService.DecryptInvitationToken(invitationToken)
	if err != nil {
		return nil, errors.New("invalid student invitation token: " + err.Error())
	}

	// Check if this invitation was already used for account creation
	used, err := studentInvitationService.IsInvitationUsedForAccountCreation(invitationToken)
	if err != nil {
		return nil, errors.New("failed to check invitation usage: " + err.Error())
	}
	if used {
		return nil, errors.New("this invitation has already been used to create an account")
	}

	// Create user with Google OAuth info
	user := models.User{
		Email:          userInfo.Email,
		FirstName:      userInfo.GivenName,
		LastName:       userInfo.FamilyName,
		IsAdmin:        false, // Invited users are not admin by default
		EmailVerified:  userInfo.VerifiedEmail,
		GoogleID:       userInfo.ID,
		AuthProvider:   "google",
		ProfilePicture: userInfo.Picture,
	}

	// Create user
	if err := config.GetDB().Create(&user).Error; err != nil {
		return nil, err
	}

	// Mark the invitation as used for account creation
	err = studentInvitationService.MarkInvitationAsUsed(invitationToken, payload, user.ID)
	if err != nil {
		// Log the error but don't fail the registration
		// The user account was already created successfully
		// TODO: Use proper logging instead of fmt.Printf
		// fmt.Printf("Warning: failed to mark invitation as used: %v\n", err)
	}

	// Redeem the student invitation to create and assign the child
	_, err = studentInvitationService.RedeemInvitation(invitationToken, user.ID)
	if err != nil {
		// If invitation redemption fails, we should clean up the created user
		// For now, we'll just return the error
		return nil, errors.New("failed to redeem student invitation: " + err.Error())
	}

	return &user, nil
}

// LinkGoogleAccount links a Google account to an existing user
func LinkGoogleAccount(userID uint, googleID, profilePicture string) error {
	result := config.GetDB().Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"google_id":       googleID,
		"auth_provider":   "google",
		"profile_picture": profilePicture,
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// MakeUserTeacher promotes a user to teacher role
func MakeUserTeacher(userID uint) (*models.User, error) {
	var user models.User
	if err := config.GetDB().First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Update the user to be a teacher
	user.IsTeacher = true
	if err := config.GetDB().Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// RemoveUserTeacher removes teacher role from a user
func RemoveUserTeacher(userID uint) (*models.User, error) {
	var user models.User
	if err := config.GetDB().First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Update the user to not be a teacher
	user.IsTeacher = false
	if err := config.GetDB().Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
