package config

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tursodatabase/libsql-client-go/libsql"
)

var DB *gorm.DB

func InitDatabase() {
	var err error
	
	// Get database URL from environment, default to local SQLite
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "file:./booktracker.db"
	}

	// Configure GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Info,
		},
	)

	var dialector gorm.Dialector

	switch {
	// Turso libSQL URL format: libsql://database-name.turso.io?authToken=token
	case strings.HasPrefix(dbURL, "libsql://"):
		fmt.Printf("Connecting to Turso libSQL database: %s\n", strings.Split(dbURL, "?")[0])
		
		// Parse the libSQL URL to extract connection details
		parsedURL, err := url.Parse(dbURL)
		if err != nil {
			log.Fatal("Failed to parse libSQL URL:", err)
		}
		
		// Extract auth token from query parameters
		authToken := parsedURL.Query().Get("authToken")
		if authToken == "" {
			log.Fatal("libSQL URL missing authToken parameter")
		}
		
		// Clean URL without auth token for connection
		baseURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)
		
		// Create libSQL connector with separate auth token
		connector, err := libsql.NewConnector(baseURL, libsql.WithAuthToken(authToken))
		if err != nil {
			log.Fatal("Failed to create libSQL connector:", err)
		}
		
		sqlDB := sql.OpenDB(connector)
		dialector = sqlite.Dialector{Conn: sqlDB}

	// Local SQLite file
	case strings.HasPrefix(dbURL, "file:"):
		filePath := strings.TrimPrefix(dbURL, "file:")
		fmt.Printf("Connecting to local SQLite database: %s\n", filePath)
		dialector = sqlite.Open(filePath)

	// Direct JDBC-style URL (convert to appropriate format)
	case strings.HasPrefix(dbURL, "jdbc:"):
		if strings.Contains(dbURL, "sqlite") {
			// Convert jdbc:sqlite:path to just path
			sqlitePath := strings.TrimPrefix(dbURL, "jdbc:sqlite:")
			fmt.Printf("Connecting to SQLite database: %s\n", sqlitePath)
			dialector = sqlite.Open(sqlitePath)
		} else if strings.Contains(dbURL, "libsql") {
			// Convert jdbc:libsql:// to libsql://
			libsqlURL := strings.Replace(dbURL, "jdbc:libsql://", "libsql://", 1)
			
			connector, err := libsql.NewConnector(libsqlURL)
			if err != nil {
				log.Fatal("Failed to create libSQL connector:", err)
			}
			
			sqlDB := sql.OpenDB(connector)
			dialector = sqlite.Dialector{Conn: sqlDB}
		} else {
			log.Fatal("Unsupported JDBC URL:", dbURL)
		}

	// Plain path (assume SQLite)
	default:
		fmt.Printf("Connecting to SQLite database: %s\n", dbURL)
		dialector = sqlite.Open(dbURL)
	}

	// Open database connection
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure SQLite pragmas (only for SQLite, not libSQL)
	if !strings.Contains(dbURL, "libsql://") {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Fatal("Failed to get underlying sql.DB:", err)
		}

		// Enable foreign keys
		_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			log.Fatal("Failed to enable foreign keys:", err)
		}

		// Enable WAL mode for better concurrency
		_, err = sqlDB.Exec("PRAGMA journal_mode = WAL")
		if err != nil {
			log.Fatal("Failed to enable WAL mode:", err)
		}
	}

	fmt.Println("Database connected successfully")
}

func GetDB() *gorm.DB {
	return DB
}