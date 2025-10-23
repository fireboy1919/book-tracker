package main

import (
	"fmt"
	"os"
	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
)

func main() {
	// Set the JWT_SECRET to match production
	os.Setenv("JWT_SECRET", "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhIjoicnciLCJpYXQiOjE3NjAxMDczOTIsImlkIjoiMGJjNDhhZTMtM2FkZi00MWQ3LTg1ODIsNjE2NDEwMjRmODk0IiwicmlkIjoiOTk5YjgzOWEtYWI2Ni00ZjY0LWIwM2EtNTE2OGJkZGMxNGI2In0.bXoSTJDJ8Sqa-tqZqLHgvd6qkd22M3KyK3vFAMhMAe36tBVt1Lm6KmZfEe_Z5dArU5z3RBa0m_P1dIXted3dCg")
	
	config.InitDatabase()
	db := config.GetDB()
	
	fmt.Println("=== Testing Auth Flow with Production Database ===")
	
	// 1. Create a new test user
	testEmail := "test-user-" + fmt.Sprintf("%d", os.Getpid()) + "@example.com"
	testPassword := "testpassword123"
	
	createReq := models.CreateUserRequest{
		Email:     testEmail,
		Password:  testPassword,
		FirstName: "Test",
		LastName:  "User",
	}
	
	fmt.Printf("1. Creating new user: %s\n", testEmail)
	user, err := services.CreateUser(createReq)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}
	fmt.Printf("   User created successfully: ID=%d\n", user.ID)
	
	// 2. Try to login with the new user
	fmt.Printf("2. Logging in with new user...\n")
	loginReq := models.LoginRequest{
		Email:    testEmail,
		Password: testPassword,
	}
	
	loginResp, err := services.Login(loginReq)
	if err != nil {
		fmt.Printf("Error logging in: %v\n", err)
		return
	}
	fmt.Printf("   Login successful! Token: %s...\n", loginResp.Token[:50])
	
	// 3. Validate the token
	fmt.Printf("3. Validating the token...\n")
	claims, err := services.ValidateToken(loginResp.Token)
	if err != nil {
		fmt.Printf("Error validating token: %v\n", err)
		return
	}
	fmt.Printf("   Token valid! Claims UserID=%d, Email=%s\n", claims.UserID, claims.Email)
	
	// 4. Try to get user by ID from claims (this is what AuthMiddleware does)
	fmt.Printf("4. Getting user by ID from claims...\n")
	foundUser, err := services.GetUserByID(claims.UserID)
	if err != nil {
		fmt.Printf("Error getting user by ID: %v\n", err)
		return
	}
	fmt.Printf("   User found! ID=%d, Email=%s\n", foundUser.ID, foundUser.Email)
	
	// 5. Clean up - delete the test user
	fmt.Printf("5. Cleaning up test user...\n")
	err = db.Delete(&models.User{}, user.ID).Error
	if err != nil {
		fmt.Printf("Error deleting test user: %v\n", err)
	} else {
		fmt.Printf("   Test user deleted successfully\n")
	}
	
	fmt.Println("=== Auth flow test completed successfully! ===")
}