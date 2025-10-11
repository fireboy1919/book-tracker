package services

import (
	"errors"

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