package main

import (
	"fmt"
	"github.com/booktracker/backend/config"
)

func main() {
	config.InitDatabase()
	db := config.GetDB()
	
	// Get the current children table structure
	var tableSQL string
	err := db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'children'").Scan(&tableSQL).Error
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Println("Current children table structure:")
	fmt.Println(tableSQL)
	fmt.Println()
	
	// Check what GORM thinks the table should look like by comparing
	hasTable := db.Migrator().HasTable("children")
	fmt.Printf("GORM thinks children table exists: %v\n", hasTable)
	
	hasClassIDColumn := db.Migrator().HasColumn("children", "class_id")
	fmt.Printf("GORM thinks class_id column exists: %v\n", hasClassIDColumn)
}