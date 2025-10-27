package main

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
)

func main() {
	log.Printf("=== Testing Student Invitation Flow ===")

	// Create a temporary in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the models
	err = db.AutoMigrate(&models.User{}, &models.Class{}, &models.ClassMembership{}, &models.Child{}, &models.UsedStudentInvitation{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create test data
	testUser := models.User{
		Email:         "teacher@test.com",
		FirstName:     "Test",
		LastName:      "Teacher",
		IsTeacher:     true,
		EmailVerified: true,
	}
	if err := db.Create(&testUser).Error; err != nil {
		log.Fatal("Failed to create test user:", err)
	}

	testClass := models.Class{
		Name:        "Test Class",
		Description: "A test class for invitation testing",
		CreatedByID: testUser.ID,
	}
	if err := db.Create(&testClass).Error; err != nil {
		log.Fatal("Failed to create test class:", err)
	}

	// Create class membership
	membership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  testUser.ID,
		Role:    "TEACHER",
	}
	if err := db.Create(&membership).Error; err != nil {
		log.Fatal("Failed to create class membership:", err)
	}

	log.Printf("Created test class %d with teacher %d", testClass.ID, testUser.ID)

	// Initialize the invitation service
	invitationService := services.NewStudentInvitationService(db)

	// Test 1: Generate invitation data (like what teacher would get)
	log.Printf("\n--- Test 1: Generate Teacher Invitation Data ---")
	invitationData, err := invitationService.GenerateTeacherInvitationData(testUser.ID, testClass.ID)
	if err != nil {
		log.Fatal("Failed to generate teacher invitation data:", err)
	}

	log.Printf("✅ Generated invitation data:")
	log.Printf("   Class ID: %d", invitationData.ClassID)
	log.Printf("   Invitation Key: %s", invitationData.InvitationKey)
	log.Printf("   Key length: %d characters", len(invitationData.InvitationKey))
	log.Printf("   Base URL: %s", invitationData.BaseURL)

	// Test 2: Generate token for specific student (like what Google Sheets would do)
	log.Printf("\n--- Test 2: Generate Student Token ---")
	studentName := "John Smith"
	payload := models.StudentInvitationPayload{
		StudentName: studentName,
		Timestamp:   time.Now().Unix(),
		ClassID:     testClass.ID, // This gets set during decryption
	}

	token, err := invitationService.EncryptInvitationDataForClass(payload, testClass.ID)
	if err != nil {
		log.Fatal("Failed to encrypt invitation data:", err)
	}

	log.Printf("✅ Generated token for student '%s':", studentName)
	log.Printf("   Token: %s", token)
	log.Printf("   Token length: %d characters", len(token))
	log.Printf("   URL would be: https://booktracker.rustyphillips.net/invite/%d/%s", testClass.ID, token)

	// Test 3: Decrypt token using class-specific method (new approach)
	log.Printf("\n--- Test 3: Decrypt Token with Class ID ---")
	decryptedPayload, err := invitationService.DecryptInvitationTokenForClass(token, testClass.ID)
	if err != nil {
		log.Fatal("Failed to decrypt invitation token:", err)
	}

	log.Printf("✅ Successfully decrypted token:")
	log.Printf("   Student Name: %s", decryptedPayload.StudentName)
	log.Printf("   Class ID: %d", decryptedPayload.ClassID)
	log.Printf("   Timestamp: %d (%s)", decryptedPayload.Timestamp, time.Unix(decryptedPayload.Timestamp, 0))

	// Verify data matches
	if decryptedPayload.StudentName != studentName {
		log.Fatal("❌ Student name mismatch!")
	}
	if decryptedPayload.ClassID != testClass.ID {
		log.Fatal("❌ Class ID mismatch!")
	}

	// Test 4: Try decrypting with wrong class ID (should fail)
	log.Printf("\n--- Test 4: Test Security - Wrong Class ID ---")
	wrongClassID := testClass.ID + 999
	_, err = invitationService.DecryptInvitationTokenForClass(token, wrongClassID)
	if err == nil {
		log.Fatal("❌ Security breach! Token decrypted with wrong class ID")
	}
	log.Printf("✅ Security working: Token correctly rejected with wrong class ID")
	log.Printf("   Error: %s", err.Error())

	// Test 5: Create a parent user and redeem invitation
	log.Printf("\n--- Test 5: Redeem Invitation ---")
	parentUser := models.User{
		Email:         "parent@test.com",
		FirstName:     "Test",
		LastName:      "Parent",
		EmailVerified: true,
	}
	if err := db.Create(&parentUser).Error; err != nil {
		log.Fatal("Failed to create parent user:", err)
	}

	child, err := invitationService.RedeemInvitationForClass(token, testClass.ID, parentUser.ID)
	if err != nil {
		log.Fatal("Failed to redeem invitation:", err)
	}

	log.Printf("✅ Successfully redeemed invitation:")
	log.Printf("   Child ID: %d", child.ID)
	log.Printf("   Child Name: %s %s", child.FirstName, child.LastName)
	log.Printf("   Owner ID: %d", child.OwnerID)
	log.Printf("   Class ID: %d", *child.ClassID)

	// Test 6: Try to redeem same token again (should fail)
	log.Printf("\n--- Test 6: Test Duplicate Redemption ---")
	_, err = invitationService.RedeemInvitationForClass(token, testClass.ID, parentUser.ID)
	if err == nil {
		log.Fatal("❌ Security breach! Token redeemed twice")
	}
	log.Printf("✅ Security working: Token correctly rejected on second use")
	log.Printf("   Error: %s", err.Error())

	// Test 7: Test token expiration (create an old token)
	log.Printf("\n--- Test 7: Test Token Expiration ---")
	oldPayload := models.StudentInvitationPayload{
		StudentName: "Old Student",
		Timestamp:   time.Now().Unix() - (31 * 24 * 3600), // 31 days ago
		ClassID:     testClass.ID,
	}

	oldToken, err := invitationService.EncryptInvitationDataForClass(oldPayload, testClass.ID)
	if err != nil {
		log.Fatal("Failed to create old token:", err)
	}

	_, err = invitationService.DecryptInvitationTokenForClass(oldToken, testClass.ID)
	if err == nil {
		log.Fatal("❌ Security breach! Expired token accepted")
	}
	log.Printf("✅ Security working: Expired token correctly rejected")
	log.Printf("   Error: %s", err.Error())

	log.Printf("\n=== All Tests Passed! ===")
	log.Printf("✅ Token generation works")
	log.Printf("✅ Token decryption works")
	log.Printf("✅ Class ID security works")
	log.Printf("✅ Token redemption works")
	log.Printf("✅ Duplicate redemption prevention works")
	log.Printf("✅ Token expiration works")
	log.Printf("\nThe invitation system is working correctly!")
}