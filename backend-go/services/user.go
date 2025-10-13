package services

import (
	"errors"
	"time"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
	"github.com/booktracker/backend-go/utils"
	"gorm.io/gorm"
)

// CreateUser creates a new user
func CreateUser(req models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	result := config.DB.Where("email = ?", req.Email).First(&existingUser)
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
		err := config.DB.Model(&models.User{}).Count(&userCount).Error
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

	result = config.DB.Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByID gets a user by ID
func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := config.DB.First(&user, id)
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
	result := config.DB.Where("email = ?", email).First(&user)
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
	result := config.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// UpdateUser updates a user
func UpdateUser(id uint, req models.UpdateUserRequest) (*models.User, error) {
	var user models.User
	result := config.DB.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	// Check if email is already taken by another user
	var existingUser models.User
	result = config.DB.Where("email = ? AND id != ?", req.Email, id).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("email already taken by another user")
	}

	// Update user
	user.Email = req.Email
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.IsAdmin = req.IsAdmin

	result = config.DB.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// DeleteUser deletes a user
func DeleteUser(id uint) error {
	result := config.DB.Delete(&models.User{}, id)
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
	result := config.DB.Where("email_verification_token = ?", token).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Check if this might be a user who is already verified
			var verifiedUser models.User
			emailResult := config.DB.Where("email_verified = ? AND email_verification_token = ''", true).First(&verifiedUser)
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

	result = config.DB.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// ResendVerificationEmail generates a new verification token for a user
func ResendVerificationEmail(email string) (*models.User, error) {
	var user models.User
	result := config.DB.Where("email = ?", email).First(&user)
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

	result = config.DB.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByVerificationToken gets a user by their verification token
func GetUserByVerificationToken(token string) (*models.User, error) {
	var user models.User
	result := config.DB.Where("email_verification_token = ?", token).First(&user)
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
	result := config.DB.Where("email = ?", email).First(&user)
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

	result = config.DB.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// ResetPassword resets a user's password using the reset token
func ResetPassword(token, newPassword string) (*models.User, error) {
	var user models.User
	result := config.DB.Where("password_reset_token = ?", token).First(&user)
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

	result = config.DB.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
