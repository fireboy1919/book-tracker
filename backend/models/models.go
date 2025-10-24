package models

import (
	"fmt"
	"log"
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
	IsTeacher              bool      `json:"isTeacher" gorm:"default:false"`
	EmailVerified          bool      `json:"emailVerified" gorm:"default:false"`
	EmailVerificationToken string    `json:"-" gorm:"index"`
	TokenExpiresAt         *time.Time `json:"-"`
	PasswordResetToken     string    `json:"-" gorm:"index"`
	PasswordResetExpiresAt *time.Time `json:"-"`
	
	// OAuth fields
	GoogleID       string    `json:"-" gorm:"index"` // Google OAuth user ID
	AuthProvider   string    `json:"authProvider" gorm:"default:'local'"` // 'local', 'google'
	ProfilePicture string    `json:"profilePicture,omitempty"` // OAuth profile picture URL
	
	// Teacher invitation system
	InvitationKey  string    `json:"-" gorm:"index"` // Encrypted key for generating stateless invitation tokens
	
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
	ClassID   *uint     `json:"classId,omitempty" gorm:"index:idx_child_class"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relationships
	Owner       User         `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Class       *Class       `json:"class,omitempty" gorm:"foreignKey:ClassID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Books       []Book       `json:"books,omitempty" gorm:"foreignKey:ChildID"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"foreignKey:ChildID"`
}

// SharedBook represents a book from Open Library that can be reused by all users
type SharedBook struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ISBN      string    `json:"isbn" gorm:"uniqueIndex"`
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
	// Track who read the book
	ReadByParent    bool   `json:"readByParent" gorm:"default:false;index:idx_book_reader"` // true if read by parent, false if read by child
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

// UsedStudentInvitation tracks student invitations that have been used for account creation
type UsedStudentInvitation struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	TokenHash     string    `json:"tokenHash" gorm:"uniqueIndex;not null"` // SHA256 hash of the invitation token
	ClassID       uint      `json:"classId" gorm:"not null;index"`
	StudentName   string    `json:"studentName" gorm:"not null"`
	CreatedUserID uint      `json:"createdUserId" gorm:"not null;index"` // User created from this invitation
	UsedAt        time.Time `json:"usedAt" gorm:"not null"`

	// Relationships
	Class       Class `json:"class,omitempty" gorm:"foreignKey:ClassID"`
	CreatedUser User  `json:"createdUser,omitempty" gorm:"foreignKey:CreatedUserID"`
}

// Class represents a classroom with reading goals
type Class struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Name              string    `json:"name" gorm:"not null"`
	Description       string    `json:"description,omitempty"`
	StudentBooksGoal  int       `json:"studentBooksGoal" gorm:"default:0"`
	OtherBooksGoal    int       `json:"otherBooksGoal" gorm:"default:0"`
	CreatedByID       uint      `json:"createdById" gorm:"not null;index:idx_class_creator"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`

	// Relationships
	CreatedBy   User              `json:"createdBy,omitempty" gorm:"foreignKey:CreatedByID"`
	Members     []ClassMembership `json:"members,omitempty" gorm:"foreignKey:ClassID"`
	Children    []Child           `json:"children,omitempty" gorm:"foreignKey:ClassID"`
}

// ClassMembership represents the relationship between users and classes
type ClassMembership struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ClassID   uint      `json:"classId" gorm:"not null;index:idx_membership_class;uniqueIndex:idx_class_user_unique"`
	UserID    uint      `json:"userId" gorm:"not null;index:idx_membership_user;uniqueIndex:idx_class_user_unique"`
	Role      string    `json:"role" gorm:"not null;check:role IN ('TEACHER', 'STUDENT')"`
	CreatedAt time.Time `json:"createdAt"`

	// Relationships
	Class Class `json:"class,omitempty" gorm:"foreignKey:ClassID"`
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Request DTOs
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	IsAdmin   bool   `json:"isAdmin"`
	IsTeacher bool   `json:"isTeacher"`
}

type CreateUserByAdminRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	IsAdmin   bool   `json:"isAdmin"`
	IsTeacher bool   `json:"isTeacher"`
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
	IsTeacher bool   `json:"isTeacher"`
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
	ReadByParent    bool   `json:"readByParent"` // true if read by parent, false if read by child
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
	ReadByParent    bool   `json:"readByParent"` // true if read by parent, false if read by child
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
	ReadByParent    bool   `json:"readByParent"` // true if read by parent, false if read by child
}

type CreatePermissionRequest struct {
	UserID         uint   `json:"userId" binding:"required"`
	ChildID        uint   `json:"childId" binding:"required"`
	PermissionType string `json:"permissionType" binding:"required,oneof=VIEW EDIT"`
}

// Class-related DTOs
type CreateClassRequest struct {
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description,omitempty"`
	StudentBooksGoal int    `json:"studentBooksGoal" binding:"min=0"`
	OtherBooksGoal   int    `json:"otherBooksGoal" binding:"min=0"`
}

type UpdateClassRequest struct {
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description,omitempty"`
	StudentBooksGoal int    `json:"studentBooksGoal" binding:"min=0"`
	OtherBooksGoal   int    `json:"otherBooksGoal" binding:"min=0"`
}

type AddClassMemberRequest struct {
	UserID uint   `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=TEACHER STUDENT"`
}

type AssignChildToClassRequest struct {
	ChildID uint `json:"childId" binding:"required"`
	ClassID uint `json:"classId" binding:"required"`
}

// Student invitation system requests
type StudentInvitationPayload struct {
	ClassID     uint   `json:"class_id"`
	StudentName string `json:"student_name"`
	Timestamp   int64  `json:"timestamp"`
}

type RedeemInvitationRequest struct {
	Token string `json:"token" binding:"required"`
}

// Response DTOs
type UserResponse struct {
	ID            uint      `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	IsAdmin       bool      `json:"isAdmin"`
	IsTeacher     bool      `json:"isTeacher"`
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
	ClassID   *uint     `json:"classId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type ChildWithBookCountResponse struct {
	ID               uint      `json:"id"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Grade            string    `json:"grade"`
	OwnerID          uint      `json:"ownerId"`
	ClassID          *uint     `json:"classId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	BookCount        int       `json:"bookCount"`        // Deprecated - total count for backward compatibility
	StudentBooksRead int       `json:"studentBooksRead"` // Books read by student
	ReadToBooksRead  int       `json:"readToBooksRead"`  // Books read to student by parent
	StudentGoal      int       `json:"studentGoal"`      // Goal for student reading
	ReadToGoal       int       `json:"readToGoal"`       // Goal for read-to books
	GoalsReached     bool      `json:"goalsReached"`     // True if both goals met
}

type ClassStudentResponse struct {
	ID               uint      `json:"id"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	OwnerID          uint      `json:"ownerId"`
	ClassID          *uint     `json:"classId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	StudentBooksRead int       `json:"studentBooksRead"`  // Books read by student
	ReadToBooksRead  int       `json:"readToBooksRead"`   // Books read to student by parent
	StudentGoal      int       `json:"studentGoal"`       // Goal for student reading
	ReadToGoal       int       `json:"readToGoal"`        // Goal for read-to books
	GoalsReached     bool      `json:"goalsReached"`      // True if both goals met
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
	ReadByParent    bool   `json:"readByParent"`
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

type ClassResponse struct {
	ID                uint      `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	StudentBooksGoal  int       `json:"studentBooksGoal"`
	OtherBooksGoal    int       `json:"otherBooksGoal"`
	CreatedByID       uint      `json:"createdById"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type ClassMembershipResponse struct {
	ID        uint          `json:"id"`
	ClassID   uint          `json:"classId"`
	UserID    uint          `json:"userId"`
	Role      string        `json:"role"`
	CreatedAt time.Time     `json:"createdAt"`
	User      *UserResponse `json:"user,omitempty"`
	Class     *ClassResponse `json:"class,omitempty"`
}

type ClassWithMembersResponse struct {
	ClassResponse
	Members  []ClassMembershipResponse `json:"members"`
	Children []ChildResponse           `json:"children"`
}

// Database migration function
func AutoMigrate(db *gorm.DB) error {
	log.Printf("=== Running GORM AutoMigrate ===")
	
	// Run GORM AutoMigrate with error handling for production
	err := db.AutoMigrate(&User{}, &SharedBook{}, &Class{}, &ClassMembership{}, &Child{}, &Book{}, &Permission{}, &PendingInvitation{}, &UsedStudentInvitation{})
	if err != nil {
		log.Printf("AutoMigrate failed: %v", err)
		// In production, don't fail completely - the tables might already exist
		// Just log the error and continue
		log.Printf("Continuing despite migration error - tables may already exist")
	} else {
		log.Printf("AutoMigrate completed successfully")
	}
	
	return nil // Always return nil to prevent serverless function panic
}

// backupDataToTempTables creates temporary tables and backs up data
func backupDataToTempTables(db *gorm.DB) error {
	log.Printf("=== Starting data backup ===")
	
	// Create temporary table for children
	err := db.Exec(`
		CREATE TABLE IF NOT EXISTS children_backup AS 
		SELECT * FROM children
	`).Error
	if err != nil {
		return fmt.Errorf("failed to backup children: %v", err)
	}
	
	// Create temporary table for books
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS books_backup AS 
		SELECT * FROM books
	`).Error
	if err != nil {
		return fmt.Errorf("failed to backup books: %v", err)
	}
	
	// Create temporary table for permissions
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS permissions_backup AS 
		SELECT * FROM permissions
	`).Error
	if err != nil {
		return fmt.Errorf("failed to backup permissions: %v", err)
	}
	
	// Create temporary table for pending_invitations if it exists
	if db.Migrator().HasTable("pending_invitations") {
		err = db.Exec(`
			CREATE TABLE IF NOT EXISTS pending_invitations_backup AS 
			SELECT * FROM pending_invitations
		`).Error
		if err != nil {
			return fmt.Errorf("failed to backup pending_invitations: %v", err)
		}
	}
	
	var childCount, bookCount, permissionCount int
	db.Raw("SELECT COUNT(*) FROM children_backup").Scan(&childCount)
	db.Raw("SELECT COUNT(*) FROM books_backup").Scan(&bookCount)
	db.Raw("SELECT COUNT(*) FROM permissions_backup").Scan(&permissionCount)
	
	log.Printf("Backed up %d children, %d books, %d permissions", childCount, bookCount, permissionCount)
	return nil
}

// clearProblematicTables drops tables and indexes that have foreign key issues
func clearProblematicTables(db *gorm.DB) error {
	log.Printf("=== Dropping problematic tables and indexes ===")
	
	// First drop any problematic indexes
	indexes := []string{"idx_child_class", "idx_child_owner"}
	for _, index := range indexes {
		err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index)).Error
		if err != nil {
			log.Printf("Warning: failed to drop index %s: %v", index, err)
		} else {
			log.Printf("Dropped index: %s", index)
		}
	}
	
	// Drop in reverse dependency order
	tables := []string{"pending_invitations", "permissions", "books", "children"}
	
	for _, table := range tables {
		if db.Migrator().HasTable(table) {
			err := db.Exec(fmt.Sprintf("DROP TABLE %s", table)).Error
			if err != nil {
				return fmt.Errorf("failed to drop table %s: %v", table, err)
			}
			log.Printf("Dropped table: %s", table)
		}
	}
	
	return nil
}

// restoreDataFromTempTables restores data back to the main tables
func restoreDataFromTempTables(db *gorm.DB) error {
	log.Printf("=== Restoring data ===")
	
	// Restore children first (no dependencies)
	err := db.Exec(`
		INSERT INTO children (id, first_name, last_name, grade, owner_id, class_id, created_at, updated_at)
		SELECT id, first_name, last_name, grade, owner_id, class_id, created_at, updated_at 
		FROM children_backup
	`).Error
	if err != nil {
		return fmt.Errorf("failed to restore children: %v", err)
	}
	
	// Restore books (depends on children)
	err = db.Exec(`
		INSERT INTO books (id, date_read, child_id, shared_book_id, custom_title, custom_author, custom_isbn, lexile_level, is_partial, partial_comment, read_by_parent, created_at, updated_at)
		SELECT id, date_read, child_id, shared_book_id, custom_title, custom_author, custom_isbn, lexile_level, is_partial, partial_comment, read_by_parent, created_at, updated_at 
		FROM books_backup
	`).Error
	if err != nil {
		return fmt.Errorf("failed to restore books: %v", err)
	}
	
	// Restore permissions (depends on children)
	err = db.Exec(`
		INSERT INTO permissions (id, user_id, child_id, permission_type, created_at)
		SELECT id, user_id, child_id, permission_type, created_at 
		FROM permissions_backup
	`).Error
	if err != nil {
		return fmt.Errorf("failed to restore permissions: %v", err)
	}
	
	// Restore pending_invitations if backup exists
	var pendingInvitationsBackupExists int
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='pending_invitations_backup'").Scan(&pendingInvitationsBackupExists)
	
	if pendingInvitationsBackupExists > 0 {
		err = db.Exec(`
			INSERT INTO pending_invitations (id, email, child_id, permission_type, invited_by_id, token, expires_at, created_at)
			SELECT id, email, child_id, permission_type, invited_by_id, token, expires_at, created_at 
			FROM pending_invitations_backup
		`).Error
		if err != nil {
			return fmt.Errorf("failed to restore pending_invitations: %v", err)
		}
	}
	
	var childCount, bookCount, permissionCount int
	db.Raw("SELECT COUNT(*) FROM children").Scan(&childCount)
	db.Raw("SELECT COUNT(*) FROM books").Scan(&bookCount)
	db.Raw("SELECT COUNT(*) FROM permissions").Scan(&permissionCount)
	
	log.Printf("Restored %d children, %d books, %d permissions", childCount, bookCount, permissionCount)
	return nil
}

// cleanupTempTables removes the temporary backup tables
func cleanupTempTables(db *gorm.DB) error {
	log.Printf("=== Cleaning up temporary tables ===")
	
	tempTables := []string{"children_backup", "books_backup", "permissions_backup", "pending_invitations_backup"}
	
	for _, table := range tempTables {
		var tableExists int
		db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableExists)
		
		if tableExists > 0 {
			err := db.Exec(fmt.Sprintf("DROP TABLE %s", table)).Error
			if err != nil {
				log.Printf("Warning: failed to drop temp table %s: %v", table, err)
			} else {
				log.Printf("Dropped temp table: %s", table)
			}
		}
	}
	
	return nil
}

// cleanAllDataForMigration cleans up data that would cause foreign key failures
func cleanAllDataForMigration(db *gorm.DB) error {
	log.Printf("=== Starting cleanAllDataForMigration ===")
	
	// Clean children data if table exists
	if db.Migrator().HasTable(&Child{}) {
		// Fix invalid class_id references  
		result1 := db.Exec("UPDATE children SET class_id = NULL WHERE class_id IS NOT NULL AND class_id NOT IN (SELECT id FROM classes)")
		if result1.RowsAffected > 0 {
			log.Printf("Fixed %d invalid class_id references", result1.RowsAffected)
		}
		
		// Fix invalid owner_id references (delete orphaned children)
		result2 := db.Exec("DELETE FROM children WHERE owner_id NOT IN (SELECT id FROM users)")
		if result2.RowsAffected > 0 {
			log.Printf("Deleted %d children with invalid owner_id references", result2.RowsAffected)
		}
	}
	
	// Clean book data if table exists
	if db.Migrator().HasTable("books") {
		// Fix invalid child_id references (delete orphaned books)
		result3 := db.Exec("DELETE FROM books WHERE child_id NOT IN (SELECT id FROM children)")
		if result3.RowsAffected > 0 {
			log.Printf("Deleted %d books with invalid child_id references", result3.RowsAffected)
		}
		
		// Fix invalid shared_book_id references
		result4 := db.Exec("UPDATE books SET shared_book_id = NULL WHERE shared_book_id IS NOT NULL AND shared_book_id NOT IN (SELECT id FROM shared_books)")
		if result4.RowsAffected > 0 {
			log.Printf("Fixed %d invalid shared_book_id references", result4.RowsAffected)
		}
	}
	
	// Clean permission data if table exists
	if db.Migrator().HasTable("permissions") {
		// Fix invalid user_id references
		result5 := db.Exec("DELETE FROM permissions WHERE user_id NOT IN (SELECT id FROM users)")
		if result5.RowsAffected > 0 {
			log.Printf("Deleted %d permissions with invalid user_id references", result5.RowsAffected)
		}
		
		// Fix invalid child_id references
		result6 := db.Exec("DELETE FROM permissions WHERE child_id NOT IN (SELECT id FROM children)")
		if result6.RowsAffected > 0 {
			log.Printf("Deleted %d permissions with invalid child_id references", result6.RowsAffected)
		}
	}
	
	log.Printf("All data cleanup completed")
	return nil
}

// ChildForMigration is a temporary struct for safe migration without foreign key constraints
type ChildForMigration struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	FirstName string    `json:"firstName" gorm:"not null"`
	LastName  string    `json:"lastName" gorm:"not null"`
	Grade     string    `json:"grade" gorm:"not null"`
	OwnerID   uint      `json:"ownerId" gorm:"not null"`
	ClassID   *uint     `json:"classId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	
	// No relationships or indexes defined - this prevents GORM from adding constraints or conflicting indexes
}

// TableName ensures this struct uses the same table as Child
func (ChildForMigration) TableName() string {
	return "children"
}

// migrateChildWithoutConstraints migrates Child model without foreign key constraints
func migrateChildWithoutConstraints(db *gorm.DB) error {
	log.Printf("=== Starting migrateChildWithoutConstraints ===")
	
	// Migrate using the constraint-free struct with explicit table name
	if err := db.AutoMigrate(&ChildForMigration{}); err != nil {
		return err
	}
	
	log.Printf("Child model migrated successfully without constraints")
	return nil
}

// Book model without foreign key constraints for migration
type BookForMigration struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	DateRead        string    `json:"dateRead" gorm:"not null"`
	ChildID         uint      `json:"childId" gorm:"not null"`
	SharedBookID    *uint     `json:"sharedBookId,omitempty"`
	CustomTitle     string    `json:"customTitle,omitempty"`
	CustomAuthor    string    `json:"customAuthor,omitempty"`
	CustomISBN      string    `json:"customIsbn,omitempty"`
	LexileLevel     string    `json:"lexileLevel,omitempty"`
	IsPartial       bool      `json:"isPartial" gorm:"default:false"`
	PartialComment  string    `json:"partialComment,omitempty"`
	ReadByParent    bool      `json:"readByParent" gorm:"default:false"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (BookForMigration) TableName() string {
	return "books"
}

// Permission model without foreign key constraints for migration
type PermissionForMigration struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"userId" gorm:"not null"`
	ChildID        uint      `json:"childId" gorm:"not null"`
	PermissionType string    `json:"permissionType" gorm:"not null;check:permission_type IN ('VIEW', 'EDIT')"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (PermissionForMigration) TableName() string {
	return "permissions"
}

// PendingInvitation model without foreign key constraints for migration
type PendingInvitationForMigration struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Email          string    `json:"email" gorm:"not null"`
	ChildID        uint      `json:"childId" gorm:"not null"`
	PermissionType string    `json:"permissionType" gorm:"not null;check:permission_type IN ('VIEW', 'EDIT')"`
	InvitedByID    uint      `json:"invitedById" gorm:"not null"`
	Token          string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt      time.Time `json:"expiresAt" gorm:"not null"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (PendingInvitationForMigration) TableName() string {
	return "pending_invitations"
}

// migrateOtherModelsWithoutConstraints migrates remaining models without foreign key constraints
func migrateOtherModelsWithoutConstraints(db *gorm.DB) error {
	log.Printf("=== Starting migrateOtherModelsWithoutConstraints ===")
	
	// Migrate all models without foreign key constraints
	if err := db.AutoMigrate(&BookForMigration{}); err != nil {
		return err
	}
	
	if err := db.AutoMigrate(&PermissionForMigration{}); err != nil {
		return err
	}
	
	if err := db.AutoMigrate(&PendingInvitationForMigration{}); err != nil {
		return err
	}
	
	log.Printf("All models migrated successfully without constraints")
	return nil
}

// addClassIDToChildren safely adds the class_id column to children table
func addClassIDToChildren(db *gorm.DB) error {
	// Check if class_id column already exists
	if db.Migrator().HasColumn(&Child{}, "class_id") {
		return nil // Column already exists, skip
	}
	
	// Add the column without foreign key constraint first
	if err := db.Exec("ALTER TABLE children ADD COLUMN class_id INTEGER").Error; err != nil {
		return err
	}
	
	// Add index for class_id if it doesn't exist
	var indexExists int
	err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'idx_child_class'").Scan(&indexExists).Error
	if err != nil {
		return err
	}
	
	if indexExists == 0 {
		if err := db.Exec("CREATE INDEX idx_child_class ON children(class_id)").Error; err != nil {
			log.Printf("Could not create index idx_child_class: %v", err)
			// Don't return error - index is not critical
		}
	}
	
	return nil
}

// addForeignKeyConstraintToChildren safely adds the foreign key constraint
func addForeignKeyConstraintToChildren(db *gorm.DB) error {
	// Clean up any invalid class_id references first
	result := db.Exec("UPDATE children SET class_id = NULL WHERE class_id IS NOT NULL AND class_id NOT IN (SELECT id FROM classes)")
	if result.Error != nil {
		return result.Error
	}
	
	// Check if the foreign key constraint already exists
	var constraintExists int
	err := db.Raw(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type = 'table' AND name = 'children' 
		AND sql LIKE '%CONSTRAINT%fk_classes_children%'
	`).Scan(&constraintExists).Error
	
	if err != nil {
		return err
	}
	
	// If constraint doesn't exist, add it using ALTER TABLE
	if constraintExists == 0 {
		err = db.Exec(`
			ALTER TABLE children 
			ADD CONSTRAINT fk_classes_children 
			FOREIGN KEY (class_id) 
			REFERENCES classes(id) 
			ON UPDATE CASCADE 
			ON DELETE SET NULL
		`).Error
		
		if err != nil {
			log.Printf("Note: Could not add foreign key constraint (will use application-level enforcement): %v", err)
			// Don't return error - constraint will be enforced at application level
		} else {
			log.Printf("Successfully added foreign key constraint to children table")
		}
	}
	
	return nil
}

// migrateChildrenTable - REMOVED to prevent data deletion
// This migration has been disabled to preserve data between deployments.
// The schema migration from 'name' to 'firstName'/'lastName' has already been applied.