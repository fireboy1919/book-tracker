package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                     uint      `json:"id" gorm:"primaryKey"`
	Email                  string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash           string    `json:"-"` // Optional for OAuth users
	FirstName              string    `json:"firstName" gorm:"not null"`
	LastName               string    `json:"lastName" gorm:"not null"`
	IsAdmin                bool      `json:"isAdmin" gorm:"default:false"`
	EmailVerified          bool      `json:"emailVerified" gorm:"default:false"`
	EmailVerificationToken string    `json:"-" gorm:"index"`
	TokenExpiresAt         *time.Time `json:"-"`
	PasswordResetToken     string    `json:"-" gorm:"index"`
	PasswordResetExpiresAt *time.Time `json:"-"`
	
	// OAuth fields
	GoogleID       string    `json:"-" gorm:"index"` // Google OAuth user ID
	AuthProvider   string    `json:"authProvider" gorm:"default:'local'"` // 'local', 'google'
	ProfilePicture string    `json:"profilePicture,omitempty"` // OAuth profile picture URL
	
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`

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
	OwnerID   uint      `json:"ownerId" gorm:"not null;index:idx_child_owner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relationships
	Owner       User         `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Books       []Book       `json:"books,omitempty" gorm:"foreignKey:ChildID"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"foreignKey:ChildID"`
}

// SharedBook represents a book from Open Library that can be reused by all users
type SharedBook struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ISBN      string    `json:"isbn" gorm:"uniqueIndex;not null"`
	Title     string    `json:"title" gorm:"not null"`
	Author    string    `json:"author" gorm:"not null"`
	CoverURL  string    `json:"coverUrl,omitempty"`
	Source    string    `json:"source" gorm:"default:'openlibrary'"` // 'openlibrary', etc.
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Book represents a reading record - links a child to either a shared book or custom book
type Book struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	DateRead     string    `json:"dateRead" gorm:"not null;index:idx_book_date"`
	ChildID      uint      `json:"childId" gorm:"not null;index:idx_book_child"`
	SharedBookID *uint     `json:"sharedBookId,omitempty" gorm:"index:idx_book_shared"` // Reference to SharedBook
	// For custom books (user-specific)
	CustomTitle  string    `json:"customTitle,omitempty" gorm:"index:idx_custom_title"`
	CustomAuthor string    `json:"customAuthor,omitempty" gorm:"index:idx_custom_author"`
	CustomISBN   string    `json:"customIsbn,omitempty"`
	LexileLevel  string    `json:"lexileLevel,omitempty"`
	// For partial books
	IsPartial       bool   `json:"isPartial" gorm:"default:false;index:idx_book_partial"`
	PartialComment  string `json:"partialComment,omitempty"` // Description of what portion was read
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	// Relationships
	Child      Child        `json:"child,omitempty" gorm:"foreignKey:ChildID"`
	SharedBook *SharedBook  `json:"sharedBook,omitempty" gorm:"foreignKey:SharedBookID"`
}

// Permission represents user permissions for children
type Permission struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"userId" gorm:"not null;index:idx_permission_user;uniqueIndex:idx_user_child_unique"`
	ChildID        uint      `json:"childId" gorm:"not null;index:idx_permission_child;uniqueIndex:idx_user_child_unique"`
	PermissionType string    `json:"permissionType" gorm:"not null;check:permission_type IN ('VIEW', 'EDIT')"`
	CreatedAt      time.Time `json:"createdAt"`

	// Relationships
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Child Child `json:"child,omitempty" gorm:"foreignKey:ChildID"`
}

// PendingInvitation represents an invitation sent to a non-registered user
type PendingInvitation struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Email          string    `json:"email" gorm:"not null;index"`
	ChildID        uint      `json:"childId" gorm:"not null;index"`
	PermissionType string    `json:"permissionType" gorm:"not null;check:permission_type IN ('VIEW', 'EDIT')"`
	InvitedByID    uint      `json:"invitedById" gorm:"not null;index"`
	Token          string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt      time.Time `json:"expiresAt" gorm:"not null"`
	CreatedAt      time.Time `json:"createdAt"`

	// Relationships
	Child     Child `json:"child,omitempty" gorm:"foreignKey:ChildID"`
	InvitedBy User  `json:"invitedBy,omitempty" gorm:"foreignKey:InvitedByID"`
}

// Request DTOs
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	IsAdmin   bool   `json:"isAdmin"`
}

type CreateUserWithInvitationRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	FirstName      string `json:"firstName" binding:"required"`
	LastName       string `json:"lastName" binding:"required"`
	InvitationToken string `json:"invitationToken" binding:"required"`
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
	ISBN         string `json:"isbn,omitempty"`
	Title        string `json:"title,omitempty"`
	Author       string `json:"author,omitempty"`
	LexileLevel  string `json:"lexileLevel,omitempty"`
	DateRead     string `json:"dateRead" binding:"required"`
	ChildID      uint   `json:"childId" binding:"required"`
	SharedBookID *uint  `json:"sharedBookId,omitempty"` // For shared books from Open Library
	IsCustomBook bool   `json:"isCustomBook"` // true for user-specific custom books
	IsPartial       bool   `json:"isPartial"` // true for partial book readings
	PartialComment  string `json:"partialComment,omitempty"` // Description of what portion was read
}

type ISBNLookupRequest struct {
	ISBN string `json:"isbn" binding:"required"`
}

type CreateCustomBookRequest struct {
	Title       string `json:"title" binding:"required"`
	Author      string `json:"author" binding:"required"`
	ISBN        string `json:"isbn,omitempty"`
	LexileLevel string `json:"lexileLevel,omitempty"`
	DateRead    string `json:"dateRead" binding:"required"`
	ChildID     uint   `json:"childId" binding:"required"`
	IsPartial       bool   `json:"isPartial"` // true for partial book readings
	PartialComment  string `json:"partialComment,omitempty"` // Description of what portion was read
}

type BookInfoResponse struct {
	ISBN        string `json:"isbn"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	LexileLevel string `json:"lexileLevel,omitempty"`
	CoverURL    string `json:"coverUrl,omitempty"`
	Found       bool   `json:"found"`
	SharedBookID *uint `json:"sharedBookId,omitempty"` // If book exists in SharedBook table
}

type InviteUserRequest struct {
	Email          string `json:"email" binding:"required,email"`
	PermissionType string `json:"permissionType" binding:"required,oneof=VIEW EDIT"`
}

type ChildPermission struct {
	ChildID        uint   `json:"childId" binding:"required"`
	PermissionType string `json:"permissionType" binding:"required,oneof=VIEW EDIT"`
}

type BulkInviteUserRequest struct {
	Email    string            `json:"email" binding:"required,email"`
	Children []ChildPermission `json:"children" binding:"required,min=1"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateBookRequest struct {
	ISBN        string `json:"isbn,omitempty"`
	Title       string `json:"title,omitempty"`
	Author      string `json:"author,omitempty"`
	LexileLevel string `json:"lexileLevel,omitempty"`
	DateRead    string `json:"dateRead" binding:"required"`
	IsPartial       bool   `json:"isPartial"` // true for partial book readings
	PartialComment  string `json:"partialComment,omitempty"` // Description of what portion was read
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
	ID           uint      `json:"id"`
	ISBN         string    `json:"isbn"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	LexileLevel  string    `json:"lexileLevel,omitempty"`
	CoverURL     string    `json:"coverUrl,omitempty"`
	DateRead     string    `json:"dateRead"`
	ChildID      uint      `json:"childId"`
	IsCustomBook bool      `json:"isCustomBook"`
	SharedBookID *uint     `json:"sharedBookId,omitempty"`
	IsPartial       bool   `json:"isPartial"`
	PartialComment  string `json:"partialComment,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
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
	// Skip the destructive migration - it's already been applied
	// TODO: Remove migrateChildrenTable function once stable
	// err := migrateChildrenTable(db)
	// if err != nil {
	// 	return err
	// }
	
	return db.AutoMigrate(&User{}, &Child{}, &SharedBook{}, &Book{}, &Permission{}, &PendingInvitation{})
}

// migrateChildrenTable - REMOVED to prevent data deletion
// This migration has been disabled to preserve data between deployments.
// The schema migration from 'name' to 'firstName'/'lastName' has already been applied.