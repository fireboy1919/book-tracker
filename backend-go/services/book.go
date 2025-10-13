package services

import (
	"errors"
	"fmt"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
	"gorm.io/gorm"
)

// CreateBook creates a new book
func CreateBook(req models.CreateBookRequest) (*models.Book, error) {
	book := models.Book{
		Title:    req.Title,
		Author:   req.Author,
		DateRead: req.DateRead,
		ChildID:  req.ChildID,
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
	result := config.DB.Preload("Child").First(&book, id)
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
	result := config.DB.Where("child_id = ?", childID).Order("date_read DESC").Find(&books)
	if result.Error != nil {
		return nil, result.Error
	}
	return books, nil
}

// GetBooksForUser gets all books that a user has permission to view
func GetBooksForUser(userID uint) ([]models.Book, error) {
	var books []models.Book
	
	// Get books for children owned by user or children user has permissions for
	result := config.DB.Raw(`
		SELECT DISTINCT b.* FROM books b 
		JOIN children c ON b.child_id = c.id 
		LEFT JOIN permissions p ON c.id = p.child_id 
		WHERE c.owner_id = ? OR p.user_id = ?
		ORDER BY b.date_read DESC
	`, userID, userID).Scan(&books)
	
	if result.Error != nil {
		return nil, result.Error
	}
	return books, nil
}

// UpdateBook updates a book
func UpdateBook(id uint, req models.UpdateBookRequest) (*models.Book, error) {
	var book models.Book
	result := config.DB.First(&book, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("book not found")
		}
		return nil, result.Error
	}

	book.Title = req.Title
	book.Author = req.Author
	book.DateRead = req.DateRead

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
	
	result := config.DB.Where("child_id = ? AND date_read >= ? AND date_read < ?", 
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