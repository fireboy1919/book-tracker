package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                     uint      `json:"id" gorm:"primaryKey"`
	Email                  string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash           string    `json:"-" gorm:"not null"`
	FirstName              string    `json:"firstName" gorm:"not null"`
	LastName               string    `json:"lastName" gorm:"not null"`
	IsAdmin                bool      `json:"isAdmin" gorm:"default:false"`
	EmailVerified          bool      `json:"emailVerified" gorm:"default:false"`
	EmailVerificationToken string    `json:"-" gorm:"index"`
	TokenExpiresAt         *time.Time `json:"-"`
	PasswordResetToken     string    `json:"-" gorm:"index"`
	PasswordResetExpiresAt *time.Time `json:"-"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`

	// Relationships
	Children    []Child      `json:"children,omitempty" gorm:"foreignKey:OwnerID"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"foreignKey:UserID"`
}

// Child represents a child in the system
type Child struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	FirstName string    `json:"firstName" gorm:"not null"`
	LastName  string    `json:"lastName" gorm:"not null"`
	Grade     string    `json:"grade" gorm:"not null"`
	OwnerID   uint      `json:"ownerId" gorm:"not null;index"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relationships
	Owner       User         `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Books       []Book       `json:"books,omitempty" gorm:"foreignKey:ChildID"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"foreignKey:ChildID"`
}

// Book represents a book that has been read
type Book struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ISBN       string    `json:"isbn"`
	Title      string    `json:"title" gorm:"not null"`
	Author     string    `json:"author" gorm:"not null"`
	LexileLevel string   `json:"lexileLevel,omitempty"` // Optional Lexile level
	DateRead   string    `json:"dateRead" gorm:"not null"`
	ChildID    uint      `json:"childId" gorm:"not null;index"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`

	// Relationships
	Child Child `json:"child,omitempty" gorm:"foreignKey:ChildID"`
}

// Permission represents user permissions for children
type Permission struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"userId" gorm:"not null;index"`
	ChildID        uint      `json:"childId" gorm:"not null;index"`
	PermissionType string    `json:"permissionType" gorm:"not null;check:permission_type IN ('VIEW', 'EDIT')"`
	CreatedAt      time.Time `json:"createdAt"`

	// Relationships
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Child Child `json:"child,omitempty" gorm:"foreignKey:ChildID"`
}

// Request DTOs
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	IsAdmin   bool   `json:"isAdmin"`
}

type UpdateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	IsAdmin   bool   `json:"isAdmin"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateChildRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Grade     string `json:"grade" binding:"required"`
}

type UpdateChildRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Grade     string `json:"grade" binding:"required"`
}

type CreateBookRequest struct {
	ISBN        string `json:"isbn" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Author      string `json:"author" binding:"required"`
	LexileLevel string `json:"lexileLevel,omitempty"`
	DateRead    string `json:"dateRead" binding:"required"`
	ChildID     uint   `json:"childId" binding:"required"`
}

type ISBNLookupRequest struct {
	ISBN string `json:"isbn" binding:"required"`
}

type BookInfoResponse struct {
	ISBN        string `json:"isbn"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	LexileLevel string `json:"lexileLevel,omitempty"`
	Found       bool   `json:"found"`
}

type InviteUserRequest struct {
	Email          string `json:"email" binding:"required,email"`
	PermissionType string `json:"permissionType" binding:"required,oneof=VIEW EDIT"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateBookRequest struct {
	ISBN        string `json:"isbn" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Author      string `json:"author" binding:"required"`
	LexileLevel string `json:"lexileLevel,omitempty"`
	DateRead    string `json:"dateRead" binding:"required"`
}

type CreatePermissionRequest struct {
	UserID         uint   `json:"userId" binding:"required"`
	ChildID        uint   `json:"childId" binding:"required"`
	PermissionType string `json:"permissionType" binding:"required,oneof=VIEW EDIT"`
}

// Response DTOs
type UserResponse struct {
	ID            uint      `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	IsAdmin       bool      `json:"isAdmin"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type ChildResponse struct {
	ID        uint      `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Grade     string    `json:"grade"`
	OwnerID   uint      `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
}

type ChildWithBookCountResponse struct {
	ID        uint      `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Grade     string    `json:"grade"`
	OwnerID   uint      `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	BookCount int       `json:"bookCount"`
}

type BookCountResponse struct {
	ChildID   uint `json:"childId"`
	BookCount int  `json:"bookCount"`
}

type BookResponse struct {
	ID          uint      `json:"id"`
	ISBN        string    `json:"isbn"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	LexileLevel string    `json:"lexileLevel,omitempty"`
	DateRead    string    `json:"dateRead"`
	ChildID     uint      `json:"childId"`
	CreatedAt   time.Time `json:"createdAt"`
}

type PermissionResponse struct {
	ID             uint          `json:"id"`
	UserID         uint          `json:"userId"`
	ChildID        uint          `json:"childId"`
	PermissionType string        `json:"permissionType"`
	CreatedAt      time.Time     `json:"createdAt"`
	User           *UserResponse `json:"user,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type ChildReportResponse struct {
	Child      ChildResponse  `json:"child"`
	Books      []BookResponse `json:"books"`
	TotalBooks int            `json:"totalBooks"`
}

type ReportResponse struct {
	Children []ChildReportResponse `json:"children"`
}

// Database migration function
func AutoMigrate(db *gorm.DB) error {
	// Perform data migration for children table if needed
	err := migrateChildrenTable(db)
	if err != nil {
		return err
	}
	
	return db.AutoMigrate(&User{}, &Child{}, &Book{}, &Permission{})
}

// migrateChildrenTable handles the migration from single 'name' field to firstName/lastName
func migrateChildrenTable(db *gorm.DB) error {
	// Check if the old schema exists (has 'name' field but no 'first_name')
	var hasName, hasFirstName bool
	
	// Check for name column
	if db.Migrator().HasColumn(&Child{}, "name") {
		hasName = true
	}
	
	// Check for first_name column  
	if db.Migrator().HasColumn(&Child{}, "first_name") {
		hasFirstName = true
	}
	
	// If we have name but not first_name, we need to migrate
	if hasName && !hasFirstName {
		// For production: Clear all data to avoid migration complexity
		// Delete in correct order to respect foreign key constraints
		
		// First delete books (they reference children)
		err := db.Exec("DELETE FROM books").Error
		if err != nil {
			return err
		}
		
		// Then delete permissions (they also reference children)
		err = db.Exec("DELETE FROM permissions").Error
		if err != nil {
			return err
		}
		
		// Finally delete children
		err = db.Exec("DELETE FROM children").Error
		if err != nil {
			return err
		}
		
		// Drop and recreate the children table
		err = db.Migrator().DropTable(&Child{})
		if err != nil {
			return err
		}
	}
	
	return nil
}