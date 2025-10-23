package main

import (
	"fmt"
	"log"
	"os"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
)

func main() {
	// Initialize database
	config.InitDatabase()
	db := config.GetDB()

	// Find the first user (oldest by creation date)
	var firstUser models.User
	result := db.Order("created_at ASC").First(&firstUser)
	if result.Error != nil {
		log.Fatal("Failed to find first user:", result.Error)
	}

	// Check if already admin
	if firstUser.IsAdmin {
		fmt.Printf("User %s (%s %s) is already an admin\n", 
			firstUser.Email, firstUser.FirstName, firstUser.LastName)
		os.Exit(0)
	}

	// Make them admin
	firstUser.IsAdmin = true
	result = db.Save(&firstUser)
	if result.Error != nil {
		log.Fatal("Failed to update user:", result.Error)
	}

	fmt.Printf("Successfully made user %s (%s %s) an admin\n", 
		firstUser.Email, firstUser.FirstName, firstUser.LastName)
}