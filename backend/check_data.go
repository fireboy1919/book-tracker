package main

import (
	"fmt"
	"github.com/booktracker/backend/config"
)

func main() {
	config.InitDatabase()
	db := config.GetDB()
	
	// Check for children with invalid owner_id
	var invalidOwners int
	db.Raw("SELECT COUNT(*) FROM children WHERE owner_id NOT IN (SELECT id FROM users)").Scan(&invalidOwners)
	fmt.Printf("Children with invalid owner_id: %d\n", invalidOwners)
	
	// Check for children with invalid class_id  
	var invalidClasses int
	db.Raw("SELECT COUNT(*) FROM children WHERE class_id IS NOT NULL AND class_id NOT IN (SELECT id FROM classes)").Scan(&invalidClasses)
	fmt.Printf("Children with invalid class_id: %d\n", invalidClasses)
	
	// Check total children and users
	var totalChildren, totalUsers int
	db.Raw("SELECT COUNT(*) FROM children").Scan(&totalChildren)
	db.Raw("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	fmt.Printf("Total children: %d, Total users: %d\n", totalChildren, totalUsers)
}