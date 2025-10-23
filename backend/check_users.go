package main

import (
	"fmt"
	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
)

func main() {
	config.InitDatabase()
	db := config.GetDB()
	
	var users []models.User
	db.Find(&users)
	
	fmt.Printf("Found %d users:\n", len(users))
	for _, u := range users {
		fmt.Printf("ID: %d, Email: %s, IsAdmin: %t\n", u.ID, u.Email, u.IsAdmin)
	}
}