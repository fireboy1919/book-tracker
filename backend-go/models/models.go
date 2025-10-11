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
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`

	// Relationships
	Children    []Child      `json:"children,omitempty" gorm:"foreignKey:OwnerID"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"foreignKey:UserID"`
}

// Child represents a child in the system
type Child struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
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
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null"`
	Author    string    `json:"author" gorm:"not null"`
	DateRead  string    `json:"dateRead" gorm:"not null"`
	ChildID   uint      `json:"childId" gorm:"not null;index"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

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
	Name  string `json:"name" binding:"required"`
	Grade string `json:"grade" binding:"required"`
}

type UpdateChildRequest struct {
	Name  string `json:"name" binding:"required"`
	Grade string `json:"grade" binding:"required"`
}

type CreateBookRequest struct {
	Title    string `json:"title" binding:"required"`
	Author   string `json:"author" binding:"required"`
	DateRead string `json:"dateRead" binding:"required"`
	ChildID  uint   `json:"childId" binding:"required"`
}

type InviteUserRequest struct {
	Email          string `json:"email" binding:"required,email"`
	PermissionType string `json:"permissionType" binding:"required,oneof=VIEW EDIT"`
}

type UpdateBookRequest struct {
	Title    string `json:"title" binding:"required"`
	Author   string `json:"author" binding:"required"`
	DateRead string `json:"dateRead" binding:"required"`
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
	Name      string    `json:"name"`
	Grade     string    `json:"grade"`
	OwnerID   uint      `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
}

type BookResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	DateRead  string    `json:"dateRead"`
	ChildID   uint      `json:"childId"`
	CreatedAt time.Time `json:"createdAt"`
}

type PermissionResponse struct {
	ID             uint      `json:"id"`
	UserID         uint      `json:"userId"`
	ChildID        uint      `json:"childId"`
	PermissionType string    `json:"permissionType"`
	CreatedAt      time.Time `json:"createdAt"`
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
	return db.AutoMigrate(&User{}, &Child{}, &Book{}, &Permission{})
}