package main

import (
	"fmt"
	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
)

func main() {
	config.InitDatabase()
	db := config.GetDB()
	
	var user models.User
	result := db.Where("email = ?", "rusty.phillips@gmail.com").First(&user)
	if result.Error != nil {
		fmt.Printf("Error finding user: %v\n", result.Error)
		return
	}
	
	fmt.Printf("Found user: %s\n", user.Email)
	fmt.Printf("Password hash exists: %t\n", user.PasswordHash != "")
	fmt.Printf("Password hash length: %d\n", len(user.PasswordHash))
	if len(user.PasswordHash) > 0 {
		fmt.Printf("Password hash starts with: %s...\n", user.PasswordHash[:10])
	}
}