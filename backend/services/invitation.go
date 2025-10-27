package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
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
	err := config.GetDB().Where("email = ? AND child_id = ?", email, childID).First(&existingInvitation).Error
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
		
		if err := config.GetDB().Save(&existingInvitation).Error; err != nil {
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

	if err := config.GetDB().Create(&invitation).Error; err != nil {
		return nil, err
	}

	return &invitation, nil
}

// GetPendingInvitationByToken retrieves a pending invitation by its token
func GetPendingInvitationByToken(token string) (*models.PendingInvitation, error) {
	var invitation models.PendingInvitation
	err := config.GetDB().Preload("Child").Preload("InvitedBy").Where("token = ? AND expires_at > ?", token, time.Now()).First(&invitation).Error
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
	if err := config.GetDB().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
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
	config.GetDB().Delete(invitation)

	return user, nil
}

// DeleteExpiredInvitations removes expired invitations (can be run periodically)
func DeleteExpiredInvitations() error {
	return config.GetDB().Where("expires_at < ?", time.Now()).Delete(&models.PendingInvitation{}).Error
}

// GetPendingInvitationsByChild gets all pending invitations for a child
func GetPendingInvitationsByChild(childID uint) ([]models.PendingInvitation, error) {
	var invitations []models.PendingInvitation
	err := config.GetDB().Preload("InvitedBy").Where("child_id = ? AND expires_at > ?", childID, time.Now()).Find(&invitations).Error
	return invitations, err
}

// CreateBulkPendingInvitation creates pending invitations for multiple children for a single email and returns the token
func CreateBulkPendingInvitation(email string, children []models.ChildPermission, invitedByID uint) (string, error) {
	// Generate a single token for all invitations for this user
	token, err := GenerateInvitationToken()
	if err != nil {
		return "", err
	}

	// Start transaction
	tx := config.GetDB().Begin()
	
	// First, delete any existing invitations for this email
	if err := tx.Where("email = ?", email).Delete(&models.PendingInvitation{}).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	
	// Create invitations for each child
	for _, childPerm := range children {
		invitation := models.PendingInvitation{
			Email:          email,
			ChildID:        childPerm.ChildID,
			PermissionType: childPerm.PermissionType,
			InvitedByID:    invitedByID,
			Token:          token, // Same token for all invitations from the same email
			ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
		}

		if err := tx.Create(&invitation).Error; err != nil {
			tx.Rollback()
			return "", err
		}
	}
	
	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		return "", err
	}
	
	return token, nil
}

// ProcessBulkInvitationRegistration creates a user account and assigns all pending permissions
func ProcessBulkInvitationRegistration(req models.CreateUserWithInvitationRequest) (*models.User, error) {
	// Get all invitations with this token
	var invitations []models.PendingInvitation
	err := config.GetDB().Where("token = ? AND expires_at > ?", req.InvitationToken, time.Now()).Find(&invitations).Error
	if err != nil {
		return nil, errors.New("invitation not found or expired")
	}
	
	if len(invitations) == 0 {
		return nil, errors.New("invitation not found or expired")
	}

	// Verify email matches invitation (all should have same email)
	if invitations[0].Email != req.Email {
		return nil, errors.New("email does not match invitation")
	}

	// Check if user already exists (shouldn't happen, but just in case)
	var existingUser models.User
	if err := config.GetDB().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Start transaction
	tx := config.GetDB().Begin()

	// Create the user
	user, err := CreateUser(models.CreateUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsAdmin:   false, // Invited users are never admins
	})
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create permissions for all children
	for _, invitation := range invitations {
		err = CreatePermission(user.ID, invitation.ChildID, invitation.PermissionType)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Delete all pending invitations for this token
	if err := tx.Where("token = ?", req.InvitationToken).Delete(&models.PendingInvitation{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetPendingInvitationsByToken gets all pending invitations by token
func GetPendingInvitationsByToken(token string) ([]models.PendingInvitation, error) {
	var invitations []models.PendingInvitation
	err := config.GetDB().Where("token = ? AND expires_at > ?", token, time.Now()).Find(&invitations).Error
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// ========== STATELESS STUDENT INVITATION SYSTEM ==========

type StudentInvitationService struct {
	DB *gorm.DB
}

func NewStudentInvitationService(db *gorm.DB) *StudentInvitationService {
	return &StudentInvitationService{DB: db}
}

// GetClassEncryptionKey returns a class-specific encryption key
func (s *StudentInvitationService) GetClassEncryptionKey(classID uint) ([]byte, error) {
	// Generate class-specific key to prevent cross-class token generation
	keyString := fmt.Sprintf("booktracker-class-%d-invitation-key-v1", classID)
	hash := sha256.Sum256([]byte(keyString))
	return hash[:], nil
}

// EncryptInvitationData encrypts student invitation data into a compact string token
// EncryptInvitationDataForClass encrypts invitation data for a specific class
func (s *StudentInvitationService) EncryptInvitationDataForClass(payload models.StudentInvitationPayload, classID uint) (string, error) {
	// Convert payload to compact pipe-delimited format: "studentName|timestamp"
	compactData := fmt.Sprintf("%s|%d", payload.StudentName, payload.Timestamp)

	// Use class-specific encryption key
	key, err := s.GetClassEncryptionKey(classID)
	if err != nil {
		return "", err
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the compact data
	ciphertext := gcm.Seal(nonce, nonce, []byte(compactData), nil)
	
	// Base64 encode the result
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptInvitationToken decrypts and validates an invitation token
func (s *StudentInvitationService) DecryptInvitationToken(token string) (*models.StudentInvitationPayload, error) {
	// Add padding to token if needed (Google Sheets removes padding)
	paddedToken := token
	for len(paddedToken)%4 != 0 {
		paddedToken += "="
	}
	
	// Decode the token
	ciphertext, err := base64.URLEncoding.DecodeString(paddedToken)
	if err != nil {
		return nil, errors.New("invalid token format")
	}

	// Get all classes to try their keys
	var classes []models.Class
	if err := s.DB.Find(&classes).Error; err != nil {
		return nil, err
	}

	for _, class := range classes {
		// Try GCM first (Go backend generated)
		if payload, err := s.tryDecryptWithClassKey(ciphertext, class.ID); err == nil {
			// Validate class ID matches
			if payload.ClassID != class.ID {
				continue // Skip if class ID doesn't match
			}
			// Validate timestamp (token expires after 30 days)
			if time.Now().Unix()-payload.Timestamp > 30*24*3600 {
				return nil, errors.New("invitation token has expired")
			}
			return payload, nil
		}
		
		// Try CBC fallback (CryptoJS generated from Google Sheets)
		if payload, err := s.tryDecryptWithClassKeyCBC(ciphertext, class.ID); err == nil {
			// Validate class ID matches
			if payload.ClassID != class.ID {
				continue // Skip if class ID doesn't match
			}
			// Validate timestamp (token expires after 30 days)
			if time.Now().Unix()-payload.Timestamp > 30*24*3600 {
				return nil, errors.New("invitation token has expired")
			}
			return payload, nil
		}
	}

	return nil, errors.New("invalid or expired invitation token")
}

// DecryptInvitationTokenForClass decrypts a token using a specific class key
func (s *StudentInvitationService) DecryptInvitationTokenForClass(token string, classID uint) (*models.StudentInvitationPayload, error) {
	// Add padding to token if needed (Google Sheets removes padding)
	paddedToken := token
	for len(paddedToken)%4 != 0 {
		paddedToken += "="
	}
	
	// Decode the token
	ciphertext, err := base64.URLEncoding.DecodeString(paddedToken)
	if err != nil {
		return nil, errors.New("invalid token format")
	}

	// Try GCM first (Go backend generated)
	if payload, err := s.tryDecryptWithClassKey(ciphertext, classID); err == nil {
		// Validate timestamp (token expires after 30 days)
		if time.Now().Unix()-payload.Timestamp > 30*24*3600 {
			return nil, errors.New("invitation token has expired")
		}
		// Set the class ID in the payload since it comes from the URL
		payload.ClassID = classID
		return payload, nil
	}
	
	// Try CBC fallback (CryptoJS generated from Google Sheets)
	if payload, err := s.tryDecryptWithClassKeyCBC(ciphertext, classID); err == nil {
		// Validate timestamp (token expires after 30 days)
		if time.Now().Unix()-payload.Timestamp > 30*24*3600 {
			return nil, errors.New("invitation token has expired")
		}
		// Set the class ID in the payload since it comes from the URL
		payload.ClassID = classID
		return payload, nil
	}

	return nil, errors.New("invalid or expired invitation token")
}

// RedeemInvitationForClass redeems an invitation for a specific class
func (s *StudentInvitationService) RedeemInvitationForClass(token string, classID uint, userID uint) (*models.Child, error) {
	// Decrypt and validate the token
	payload, err := s.DecryptInvitationTokenForClass(token, classID)
	if err != nil {
		return nil, err
	}

	// Check if this invitation was already used
	tokenHash := s.hashToken(token)
	var existingInvitation models.UsedStudentInvitation
	if err := s.DB.Where("token_hash = ?", tokenHash).First(&existingInvitation).Error; err == nil {
		return nil, errors.New("invitation token has already been used")
	}

	// Get the class to verify it exists
	var class models.Class
	if err := s.DB.First(&class, classID).Error; err != nil {
		return nil, errors.New("class not found")
	}

	// Check if user already has a child with this name in this class
	var existingChild models.Child
	if err := s.DB.Where("owner_id = ? AND first_name = ? AND last_name = ? AND class_id = ?", 
		userID, payload.StudentName, "", classID).First(&existingChild).Error; err == nil {
		return nil, errors.New("you already have a student with this name in this class")
	}

	// Check if user already has a child with this name (update their class)
	if err := s.DB.Where("owner_id = ? AND first_name = ?", userID, payload.StudentName).First(&existingChild).Error; err == nil {
		// Update existing child's class
		if existingChild.ClassID == nil || *existingChild.ClassID != classID {
			existingChild.ClassID = &classID
			if err := s.DB.Save(&existingChild).Error; err != nil {
				return nil, err
			}
		}
	} else {
		// Create new child
		existingChild = models.Child{
			FirstName: payload.StudentName,
			LastName:  "", // We only have the full name from the invitation
			Grade:     "Unknown", // Will need to be updated by parent
			OwnerID:   userID,
			ClassID:   &classID,
		}
		
		if err := s.DB.Create(&existingChild).Error; err != nil {
			return nil, err
		}
	}

	// Mark invitation as used
	usedInvitation := models.UsedStudentInvitation{
		TokenHash:     tokenHash,
		ClassID:       classID,
		StudentName:   payload.StudentName,
		CreatedUserID: userID,
		UsedAt:        time.Now(),
	}
	
	if err := s.DB.Create(&usedInvitation).Error; err != nil {
		return nil, err
	}

	return &existingChild, nil
}

// hashToken creates a SHA256 hash of the token for storage
func (s *StudentInvitationService) hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Helper function to decrypt with a specific class key (GCM mode)
func (s *StudentInvitationService) tryDecryptWithClassKey(ciphertext []byte, classID uint) (*models.StudentInvitationPayload, error) {
	// Get class-specific key
	key, err := s.GetClassEncryptionKey(classID)
	if err != nil {
		return nil, err
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Check minimum length
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// Parse compact format: "classId|studentName|timestamp"
	payload, err := s.parseCompactFormat(string(plaintext))
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// tryDecryptWithKeyCBC - fallback method to decrypt CryptoJS CBC encrypted tokens
func (s *StudentInvitationService) tryDecryptWithKeyCBC(ciphertext []byte, teacherKey string) (*models.StudentInvitationPayload, error) {
	// Decode the teacher's key
	key, err := base64.URLEncoding.DecodeString(teacherKey)
	if err != nil {
		return nil, err
	}

	// For CBC, we need 16-byte IV + ciphertext
	if len(ciphertext) < 16 {
		return nil, errors.New("ciphertext too short for CBC")
	}

	// Extract IV and ciphertext
	iv := ciphertext[:16]
	encryptedData := ciphertext[16:]

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create CBC mode
	if len(encryptedData)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(encryptedData))
	mode.CryptBlocks(plaintext, encryptedData)

	// Remove PKCS7 padding
	plaintext, err = removePKCS7Padding(plaintext)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var payload models.StudentInvitationPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

// removePKCS7Padding removes PKCS7 padding from decrypted data
func removePKCS7Padding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("invalid padding")
	}

	padLength := int(data[length-1])
	if padLength > length || padLength == 0 {
		return nil, errors.New("invalid padding")
	}

	// Check that all padding bytes are the same
	for i := length - padLength; i < length; i++ {
		if data[i] != byte(padLength) {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:length-padLength], nil
}

// tryDecryptWithClassKeyCBC - class-specific CBC decryption for Google Sheets compatibility
func (s *StudentInvitationService) tryDecryptWithClassKeyCBC(ciphertext []byte, classID uint) (*models.StudentInvitationPayload, error) {
	// Get class-specific key
	key, err := s.GetClassEncryptionKey(classID)
	if err != nil {
		return nil, err
	}

	// For CBC, we need 16-byte IV + ciphertext
	if len(ciphertext) < 16 {
		return nil, errors.New("ciphertext too short for CBC")
	}

	// Extract IV and ciphertext
	iv := ciphertext[:16]
	encryptedData := ciphertext[16:]

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create CBC mode
	if len(encryptedData)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(encryptedData))
	mode.CryptBlocks(plaintext, encryptedData)

	// Remove PKCS7 padding
	plaintext, err = removePKCS7Padding(plaintext)
	if err != nil {
		return nil, err
	}

	// Parse compact format: "classId|studentName|timestamp"
	payload, err := s.parseCompactFormat(string(plaintext))
	if err != nil {
		return nil, err
	}

	return payload, nil
}


// parseCompactFormat parses the pipe-delimited compact format: "classId|studentName|timestamp"
func (s *StudentInvitationService) parseCompactFormat(compactData string) (*models.StudentInvitationPayload, error) {
	parts := strings.Split(compactData, "|")
	if len(parts) != 2 {
		return nil, errors.New("invalid compact format")
	}

	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, errors.New("invalid timestamp")
	}

	return &models.StudentInvitationPayload{
		StudentName: parts[0],
		Timestamp:   timestamp,
	}, nil
}

// hashInvitationToken creates a SHA256 hash of the invitation token for tracking usage
func (s *StudentInvitationService) hashInvitationToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// IsInvitationUsedForAccountCreation checks if an invitation token was already used to create an account
func (s *StudentInvitationService) IsInvitationUsedForAccountCreation(token string) (bool, error) {
	tokenHash := s.hashInvitationToken(token)
	
	var usedInvitation models.UsedStudentInvitation
	err := s.DB.Where("token_hash = ?", tokenHash).First(&usedInvitation).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // Not used
		}
		return false, err // Database error
	}
	
	return true, nil // Already used
}

// MarkInvitationAsUsed records that an invitation token was used to create an account
func (s *StudentInvitationService) MarkInvitationAsUsed(token string, payload *models.StudentInvitationPayload, createdUserID uint) error {
	tokenHash := s.hashInvitationToken(token)
	
	usedInvitation := models.UsedStudentInvitation{
		TokenHash:     tokenHash,
		ClassID:       payload.ClassID,
		StudentName:   payload.StudentName,
		CreatedUserID: createdUserID,
		UsedAt:        time.Now(),
	}
	
	return s.DB.Create(&usedInvitation).Error
}

// RedeemInvitation processes an invitation token and creates/assigns a child
func (s *StudentInvitationService) RedeemInvitation(token string, parentUserID uint) (*models.ChildResponse, error) {
	// Decrypt and validate the token
	payload, err := s.DecryptInvitationToken(token)
	if err != nil {
		return nil, err
	}

	// Verify the class still exists
	var class models.Class
	if err := s.DB.First(&class, payload.ClassID).Error; err != nil {
		return nil, errors.New("class not found")
	}

	// Check if parent already has a child with this name (without grade matching since we eliminated it)
	var existingChild models.Child
	err = s.DB.Where("owner_id = ? AND first_name || ' ' || last_name = ?", 
		parentUserID, payload.StudentName).First(&existingChild).Error
	
	if err == nil {
		// Child exists, assign to class if not already assigned
		if existingChild.ClassID == nil || *existingChild.ClassID != payload.ClassID {
			existingChild.ClassID = &payload.ClassID
			if err := s.DB.Save(&existingChild).Error; err != nil {
				return nil, err
			}
		}
		
		return &models.ChildResponse{
			ID:        existingChild.ID,
			FirstName: existingChild.FirstName,
			LastName:  existingChild.LastName,
			Grade:     existingChild.Grade,
			OwnerID:   existingChild.OwnerID,
			ClassID:   existingChild.ClassID,
			CreatedAt: existingChild.CreatedAt,
		}, nil
	}

	// Child doesn't exist, create new one - we need a default grade since it's required
	names := parseFullName(payload.StudentName)
	newChild := models.Child{
		FirstName: names.FirstName,
		LastName:  names.LastName,
		Grade:     class.Name, // Use class name as default grade since we don't have grade in payload
		OwnerID:   parentUserID,
		ClassID:   &payload.ClassID,
	}

	if err := s.DB.Create(&newChild).Error; err != nil {
		return nil, err
	}

	return &models.ChildResponse{
		ID:        newChild.ID,
		FirstName: newChild.FirstName,
		LastName:  newChild.LastName,
		Grade:     newChild.Grade,
		OwnerID:   newChild.OwnerID,
		ClassID:   newChild.ClassID,
		CreatedAt: newChild.CreatedAt,
	}, nil
}

// Helper to parse full name into first/last name
type ParsedName struct {
	FirstName string
	LastName  string
}

func parseFullName(fullName string) ParsedName {
	// Simple parsing - can be enhanced later
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return ParsedName{FirstName: "Unknown", LastName: "Student"}
	} else if len(parts) == 1 {
		return ParsedName{FirstName: parts[0], LastName: ""}
	} else {
		firstName := parts[0]
		lastName := strings.Join(parts[1:], " ")
		return ParsedName{FirstName: firstName, LastName: lastName}
	}
}

// GenerateTeacherInvitationData creates the data needed for Google Sheets
func (s *StudentInvitationService) GenerateTeacherInvitationData(teacherID uint, classID uint) (*TeacherInvitationData, error) {
	// Get class-specific encryption key
	key, err := s.GetClassEncryptionKey(classID)
	if err != nil {
		return nil, err
	}

	// Get class information
	var class models.Class
	if err := s.DB.First(&class, classID).Error; err != nil {
		return nil, err
	}

	// Create compound key: "classId|hexKey"
	hexKey := hex.EncodeToString(key)
	compoundData := fmt.Sprintf("%d|%s", classID, hexKey)
	
	// Base64 encode the compound key for easy copying
	compoundKey := base64.StdEncoding.EncodeToString([]byte(compoundData))

	return &TeacherInvitationData{
		TeacherID:     teacherID,
		ClassID:       classID,
		ClassName:     class.Name,
		InvitationKey: compoundKey, // Now contains both class ID and encryption key
		BaseURL:       "https://booktracker.rustyphillips.net/invite/",
	}, nil
}

type TeacherInvitationData struct {
	TeacherID     uint   `json:"teacher_id"`
	ClassID       uint   `json:"class_id"`
	ClassName     string `json:"class_name"`
	InvitationKey string `json:"invitation_key"`
	BaseURL       string `json:"base_url"`
}