package main

import (
	"fmt"
	"log"
	"os"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run emergency_admin.go <email>")
		fmt.Println("This will make the user with the specified email an admin")
		os.Exit(1)
	}

	email := os.Args[1]

	// Initialize database
	config.InitDatabase()
	db := config.GetDB()

	// Find user by email
	var user models.User
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		log.Fatal("Failed to find user with email:", email, result.Error)
	}

	// Check if already admin
	if user.IsAdmin {
		fmt.Printf("User %s (%s %s) is already an admin\n", 
			user.Email, user.FirstName, user.LastName)
		os.Exit(0)
	}

	// Make them admin
	user.IsAdmin = true
	result = db.Save(&user)
	if result.Error != nil {
		log.Fatal("Failed to update user:", result.Error)
	}

	fmt.Printf("Successfully made user %s (%s %s) an admin\n", 
		user.Email, user.FirstName, user.LastName)
}