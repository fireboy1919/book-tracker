package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
	"gorm.io/gorm"
)

// GenerateInvitationToken generates a secure random token for invitations
func GenerateInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreatePendingInvitation creates a pending invitation for a non-registered user
func CreatePendingInvitation(email string, childID uint, permissionType string, invitedByID uint) (*models.PendingInvitation, error) {
	// Check if there's already a pending invitation for this email and child
	var existingInvitation models.PendingInvitation
	err := config.DB.Where("email = ? AND child_id = ?", email, childID).First(&existingInvitation).Error
	if err == nil {
		// Update existing invitation with new permission type and extend expiration
		token, err := GenerateInvitationToken()
		if err != nil {
			return nil, err
		}
		
		existingInvitation.PermissionType = permissionType
		existingInvitation.InvitedByID = invitedByID
		existingInvitation.Token = token
		existingInvitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days
		
		if err := config.DB.Save(&existingInvitation).Error; err != nil {
			return nil, err
		}
		return &existingInvitation, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new invitation
	token, err := GenerateInvitationToken()
	if err != nil {
		return nil, err
	}

	invitation := models.PendingInvitation{
		Email:          email,
		ChildID:        childID,
		PermissionType: permissionType,
		InvitedByID:    invitedByID,
		Token:          token,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := config.DB.Create(&invitation).Error; err != nil {
		return nil, err
	}

	return &invitation, nil
}

// GetPendingInvitationByToken retrieves a pending invitation by its token
func GetPendingInvitationByToken(token string) (*models.PendingInvitation, error) {
	var invitation models.PendingInvitation
	err := config.DB.Preload("Child").Preload("InvitedBy").Where("token = ? AND expires_at > ?", token, time.Now()).First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found or expired")
		}
		return nil, err
	}
	return &invitation, nil
}

// ProcessInvitationRegistration creates a user account and assigns permissions based on invitation
func ProcessInvitationRegistration(req models.CreateUserWithInvitationRequest) (*models.User, error) {
	// Get the invitation
	invitation, err := GetPendingInvitationByToken(req.InvitationToken)
	if err != nil {
		return nil, err
	}

	// Verify email matches invitation
	if invitation.Email != req.Email {
		return nil, errors.New("email does not match invitation")
	}

	// Check if user already exists (shouldn't happen, but just in case)
	var existingUser models.User
	if err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Create the user
	user, err := CreateUser(models.CreateUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsAdmin:   false, // Invited users are never admins
	})
	if err != nil {
		return nil, err
	}

	// Create the permission
	err = CreatePermission(user.ID, invitation.ChildID, invitation.PermissionType)
	if err != nil {
		// If permission creation fails, we should probably clean up the user
		// But for simplicity, we'll just return the error
		return nil, err
	}

	// Delete the pending invitation since it's been processed
	config.DB.Delete(invitation)

	return user, nil
}

// DeleteExpiredInvitations removes expired invitations (can be run periodically)
func DeleteExpiredInvitations() error {
	return config.DB.Where("expires_at < ?", time.Now()).Delete(&models.PendingInvitation{}).Error
}

// GetPendingInvitationsByChild gets all pending invitations for a child
func GetPendingInvitationsByChild(childID uint) ([]models.PendingInvitation, error) {
	var invitations []models.PendingInvitation
	err := config.DB.Preload("InvitedBy").Where("child_id = ? AND expires_at > ?", childID, time.Now()).Find(&invitations).Error
	return invitations, err
}