package config

import (
	"log"

	"github.com/booktracker/api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var TestDB *gorm.DB

// SetupTestDatabase initializes an in-memory SQLite database for testing
func SetupTestDatabase() *gorm.DB {
	var err error
	
	// Use in-memory SQLite database for tests
	TestDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable logging in tests
	})
	
	if err != nil {
		log.Fatal("Failed to connect to test database:", err)
	}

	// Auto-migrate the schema
	err = models.AutoMigrate(TestDB)
	if err != nil {
		log.Fatal("Failed to migrate test database:", err)
	}

	// Enable foreign keys
	sqlDB, err := TestDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Failed to enable foreign keys:", err)
	}

	return TestDB
}

// CleanupTestDatabase clears all data from the test database
func CleanupTestDatabase() {
	if TestDB == nil {
		return
	}

	// Delete data in order to respect foreign key constraints
	TestDB.Exec("DELETE FROM books")
	TestDB.Exec("DELETE FROM permissions")
	TestDB.Exec("DELETE FROM children")
	TestDB.Exec("DELETE FROM users")
	
	// Reset auto-increment counters
	TestDB.Exec("DELETE FROM sqlite_sequence")
}

// GetTestDB returns the test database instance
func GetTestDB() *gorm.DB {
	return TestDB
}