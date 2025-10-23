package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Minimal test of the compact invitation system
type TestPayload struct {
	ClassID     uint   `json:"class_id"`
	StudentName string `json:"student_name"`
	Timestamp   int64  `json:"timestamp"`
}

func getClassKey(classID uint) []byte {
	keyString := fmt.Sprintf("booktracker-class-%d-invitation-key-v1", classID)
	hash := sha256.Sum256([]byte(keyString))
	return hash[:]
}

func encryptCompact(classID uint, studentName string) (string, error) {
	// Create compact format
	timestamp := time.Now().Unix()
	compactData := fmt.Sprintf("%d|%s|%d", classID, studentName, timestamp)
	
	// Get class key
	key := getClassKey(classID)
	
	// Encrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(compactData), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func decryptCompact(token string, classID uint) (*TestPayload, error) {
	// Decode token
	ciphertext, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	
	// Get class key
	key := getClassKey(classID)
	
	// Decrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	
	// Parse compact format
	parts := strings.Split(string(plaintext), "|")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid format")
	}
	
	parsedClassID, _ := strconv.ParseUint(parts[0], 10, 32)
	timestamp, _ := strconv.ParseInt(parts[2], 10, 64)
	
	return &TestPayload{
		ClassID:     uint(parsedClassID),
		StudentName: parts[1],
		Timestamp:   timestamp,
	}, nil
}

func main() {
	fmt.Println("🔒 Testing Compact Invitation System")
	
	// Test data
	classID := uint(123)
	studentName := "John Smith"
	
	// Generate token
	fmt.Printf("\n📝 Generating token for Class %d, Student: %s\n", classID, studentName)
	token, err := encryptCompact(classID, studentName)
	if err != nil {
		fmt.Printf("❌ Encryption failed: %v\n", err)
		return
	}
	
	fmt.Printf("🎯 Generated token: %s\n", token)
	fmt.Printf("📏 Token length: %d characters\n", len(token))
	
	// Test decryption with correct class
	fmt.Printf("\n🔓 Testing decryption with correct class ID (%d)...\n", classID)
	payload, err := decryptCompact(token, classID)
	if err != nil {
		fmt.Printf("❌ Decryption failed: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Decryption successful!\n")
	fmt.Printf("   Class ID: %d\n", payload.ClassID)
	fmt.Printf("   Student: %s\n", payload.StudentName) 
	fmt.Printf("   Timestamp: %d\n", payload.Timestamp)
	
	// Test decryption with wrong class (should fail)
	fmt.Printf("\n🚫 Testing decryption with wrong class ID (999)...\n")
	_, err = decryptCompact(token, 999)
	if err != nil {
		fmt.Printf("✅ Correctly rejected wrong class: %v\n", err)
	} else {
		fmt.Printf("❌ Security issue: accepted wrong class!\n")
	}
	
	fmt.Printf("\n🎉 Compact invitation system working correctly!\n")
	
	// Test different students
	fmt.Printf("\n📋 Testing with different student names:\n")
	students := []string{"Sarah Johnson", "Mike Brown", "Emma Davis"}
	
	for _, student := range students {
		token, err := encryptCompact(classID, student)
		if err != nil {
			fmt.Printf("❌ Failed for %s: %v\n", student, err)
			continue
		}
		
		payload, err := decryptCompact(token, classID)
		if err != nil {
			fmt.Printf("❌ Decrypt failed for %s: %v\n", student, err)
			continue
		}
		
		fmt.Printf("✅ %s → %d chars → %s\n", student, len(token), payload.StudentName)
	}
}