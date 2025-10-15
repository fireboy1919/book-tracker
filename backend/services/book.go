package services

import (
	"errors"
	"fmt"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"gorm.io/gorm"
)

// CreateBook creates a new book reading record
func CreateBook(req models.CreateBookRequest) (*models.Book, error) {
	// For partial books, we allow duplicates since they represent different portions
	if !req.IsPartial {
		// Check for duplicate reading record for this child (only for non-partial books)
		var existingBook models.Book
		var duplicateQuery *gorm.DB
		
		if req.SharedBookID != nil {
			// Check if child has already read this shared book (non-partial)
			duplicateQuery = config.DB.Where("child_id = ? AND shared_book_id = ? AND is_partial = ?", req.ChildID, *req.SharedBookID, false)
		} else if req.IsCustomBook {
			// Check if child has already read this custom book (by title + author, non-partial)
			duplicateQuery = config.DB.Where("child_id = ? AND custom_title = ? AND custom_author = ? AND is_partial = ?", 
				req.ChildID, req.Title, req.Author, false)
		} else {
			return nil, errors.New("invalid book request: must specify either shared book ID or custom book")
		}
		
		if err := duplicateQuery.First(&existingBook).Error; err == nil {
			return nil, errors.New("child has already read this book")
		}
	}

	book := models.Book{
		DateRead:       req.DateRead,
		ChildID:        req.ChildID,
		SharedBookID:   req.SharedBookID,
		LexileLevel:    req.LexileLevel,
		IsPartial:      req.IsPartial,
		PartialComment: req.PartialComment,
	}
	
	// Set custom book fields if this is a custom book
	if req.IsCustomBook {
		book.CustomTitle = req.Title
		book.CustomAuthor = req.Author
		book.CustomISBN = req.ISBN
	}

	result := config.DB.Create(&book)
	if result.Error != nil {
		return nil, result.Error
	}

	return &book, nil
}

// GetBookByID gets a book by ID
func GetBookByID(id uint) (*models.Book, error) {
	var book models.Book
	result := config.DB.Preload("Child").Preload("SharedBook").First(&book, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("book not found")
		}
		return nil, result.Error
	}
	return &book, nil
}

// GetBooksByChild gets all books for a child
func GetBooksByChild(childID uint) ([]models.Book, error) {
	var books []models.Book
	result := config.DB.Preload("SharedBook").Where("child_id = ?", childID).Order("date_read DESC").Find(&books)
	if result.Error != nil {
		return nil, result.Error
	}
	return books, nil
}

// GetBooksForUser gets all books that a user has permission to view
func GetBooksForUser(userID uint) ([]models.Book, error) {
	var books []models.Book
	
	// Get books for children owned by user or children user has permissions for
	result := config.DB.Preload("SharedBook").Raw(`
		SELECT DISTINCT b.* FROM books b 
		JOIN children c ON b.child_id = c.id 
		LEFT JOIN permissions p ON c.id = p.child_id 
		WHERE c.owner_id = ? OR p.user_id = ?
		ORDER BY b.date_read DESC
	`, userID, userID).Find(&books)
	
	if result.Error != nil {
		return nil, result.Error
	}
	return books, nil
}

// UpdateBook updates a book reading record
func UpdateBook(id uint, req models.UpdateBookRequest) (*models.Book, error) {
	var book models.Book
	result := config.DB.Preload("SharedBook").First(&book, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("book not found")
		}
		return nil, result.Error
	}

	// Allow updating date read, lexile level, and partial info
	book.DateRead = req.DateRead
	book.LexileLevel = req.LexileLevel
	book.IsPartial = req.IsPartial
	book.PartialComment = req.PartialComment
	
	// If this is a custom book, allow updating the book details
	if book.SharedBookID == nil {
		if req.Title != "" {
			book.CustomTitle = req.Title
		}
		if req.Author != "" {
			book.CustomAuthor = req.Author
		}
		if req.ISBN != "" {
			book.CustomISBN = req.ISBN
		}
	}

	result = config.DB.Save(&book)
	if result.Error != nil {
		return nil, result.Error
	}

	return &book, nil
}

// DeleteBook deletes a book
func DeleteBook(id uint) error {
	result := config.DB.Delete(&models.Book{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("book not found")
	}
	return nil
}

// GetBooksByChildAndMonth gets books for a child in a specific month/year
func GetBooksByChildAndMonth(childID uint, year int, month int) ([]models.Book, error) {
	var books []models.Book
	
	// Create start and end dates for the month
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	
	// Calculate end date (first day of next month)
	endYear := year
	endMonth := month + 1
	if endMonth > 12 {
		endMonth = 1
		endYear++
	}
	endDate := fmt.Sprintf("%d-%02d-01", endYear, endMonth)
	
	result := config.DB.Preload("SharedBook").Where("child_id = ? AND date_read >= ? AND date_read < ?", 
		childID, startDate, endDate).Order("date_read DESC").Find(&books)
	
	return books, result.Error
}

// GetBookCountByChildAndMonth gets the count of books for a child in a specific month/year
func GetBookCountByChildAndMonth(childID uint, year int, month int) (int, error) {
	var count int64
	
	// Create start and end dates for the month
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	
	// Calculate end date (first day of next month)
	endYear := year
	endMonth := month + 1
	if endMonth > 12 {
		endMonth = 1
		endYear++
	}
	endDate := fmt.Sprintf("%d-%02d-01", endYear, endMonth)
	
	result := config.DB.Model(&models.Book{}).Where("child_id = ? AND date_read >= ? AND date_read < ?", 
		childID, startDate, endDate).Count(&count)
	
	return int(count), result.Error
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return config.DB
}

// CreateCustomBook creates a custom book reading record
func CreateCustomBook(req models.CreateCustomBookRequest) (*models.Book, error) {
	// For partial books, we allow duplicates since they represent different portions
	if !req.IsPartial {
		// Check for duplicate custom book for this child (only for non-partial books)
		var existingBook models.Book
		if err := config.DB.Where("child_id = ? AND custom_title = ? AND custom_author = ? AND is_partial = ?", 
			req.ChildID, req.Title, req.Author, false).First(&existingBook).Error; err == nil {
			return nil, errors.New("child has already read this book")
		}
	}

	book := models.Book{
		DateRead:       req.DateRead,
		ChildID:        req.ChildID,
		CustomTitle:    req.Title,
		CustomAuthor:   req.Author,
		CustomISBN:     req.ISBN,
		LexileLevel:    req.LexileLevel,
		IsPartial:      req.IsPartial,
		PartialComment: req.PartialComment,
	}

	result := config.DB.Create(&book)
	if result.Error != nil {
		return nil, result.Error
	}

	return &book, nil
}