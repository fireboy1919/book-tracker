package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateVerificationToken generates a secure random token for email verification
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 64 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetTokenExpiration returns the expiration time for email verification tokens
func GetTokenExpiration() time.Time {
	return time.Now().Add(24 * time.Hour) // 24 hours from now
}

// GetInvitationExpiration returns the expiration time for invitation tokens
func GetInvitationExpiration() time.Time {
	return time.Now().Add(7 * 24 * time.Hour) // 7 days from now
}

// IsTokenExpired checks if a token has expired
func IsTokenExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return true
	}
	return time.Now().After(*expiresAt)
}