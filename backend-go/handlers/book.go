package handlers

import (
	"net/http"
	"strconv"

	"github.com/booktracker/backend-go/middleware"
	"github.com/booktracker/backend-go/models"
	"github.com/booktracker/backend-go/services"
	"github.com/gin-gonic/gin"
)

// CreateBook handles creating a new book
func CreateBook(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	var req models.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check permission to edit the child
	hasPermission, err := services.CheckChildPermission(userID, req.ChildID, "EDIT")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to check permission: " + err.Error(),
		})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	book, err := services.CreateBook(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to create book: " + err.Error(),
		})
		return
	}

	bookResponse := models.BookResponse{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		DateRead:  book.DateRead,
		ChildID:   book.ChildID,
		CreatedAt: book.CreatedAt,
	}

	c.JSON(http.StatusCreated, bookResponse)
}

// GetBooks handles getting books for current user
func GetBooks(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	// Check if filtering by child ID
	childIDParam := c.Query("childId")
	if childIDParam != "" {
		childID, err := strconv.ParseUint(childIDParam, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Message: "Invalid child ID",
			})
			return
		}

		// Check permission
		hasPermission, err := services.CheckChildPermission(userID, uint(childID), "VIEW")
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to check permission: " + err.Error(),
			})
			return
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Message: "Access denied",
			})
			return
		}

		books, err := services.GetBooksByChild(uint(childID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to get books: " + err.Error(),
			})
			return
		}

		var bookResponses []models.BookResponse
		for _, book := range books {
			bookResponses = append(bookResponses, models.BookResponse{
				ID:        book.ID,
				Title:     book.Title,
				Author:    book.Author,
				DateRead:  book.DateRead,
				ChildID:   book.ChildID,
				CreatedAt: book.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, bookResponses)
		return
	}

	// Get all books for user
	books, err := services.GetBooksForUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to get books: " + err.Error(),
		})
		return
	}

	var bookResponses []models.BookResponse
	for _, book := range books {
		bookResponses = append(bookResponses, models.BookResponse{
			ID:        book.ID,
			Title:     book.Title,
			Author:    book.Author,
			DateRead:  book.DateRead,
			ChildID:   book.ChildID,
			CreatedAt: book.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, bookResponses)
}

// GetBookByID handles getting a book by ID
func GetBookByID(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid book ID",
		})
		return
	}

	book, err := services.GetBookByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	// Check permission
	hasPermission, err := services.CheckChildPermission(userID, book.ChildID, "VIEW")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to check permission: " + err.Error(),
		})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	bookResponse := models.BookResponse{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		DateRead:  book.DateRead,
		ChildID:   book.ChildID,
		CreatedAt: book.CreatedAt,
	}

	c.JSON(http.StatusOK, bookResponse)
}

// UpdateBook handles updating a book
func UpdateBook(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid book ID",
		})
		return
	}

	// Get book to check child permission
	book, err := services.GetBookByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	// Check permission
	hasPermission, err := services.CheckChildPermission(userID, book.ChildID, "EDIT")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to check permission: " + err.Error(),
		})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	var req models.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	updatedBook, err := services.UpdateBook(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	bookResponse := models.BookResponse{
		ID:        updatedBook.ID,
		Title:     updatedBook.Title,
		Author:    updatedBook.Author,
		DateRead:  updatedBook.DateRead,
		ChildID:   updatedBook.ChildID,
		CreatedAt: updatedBook.CreatedAt,
	}

	c.JSON(http.StatusOK, bookResponse)
}

// DeleteBook handles deleting a book
func DeleteBook(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid book ID",
		})
		return
	}

	// Get book to check child permission
	book, err := services.GetBookByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	// Check permission
	hasPermission, err := services.CheckChildPermission(userID, book.ChildID, "EDIT")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to check permission: " + err.Error(),
		})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	err = services.DeleteBook(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// CreateBookForChild handles creating a book for a specific child
func CreateBookForChild(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	childIDParam := c.Param("childId")
	childID, err := strconv.ParseUint(childIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid child ID",
		})
		return
	}

	var req models.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Override childId from URL
	req.ChildID = uint(childID)

	// Check permission to edit the child
	hasPermission, err := services.CheckChildPermission(userID, req.ChildID, "EDIT")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to check permission: " + err.Error(),
		})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	book, err := services.CreateBook(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to create book: " + err.Error(),
		})
		return
	}

	bookResponse := models.BookResponse{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		DateRead:  book.DateRead,
		ChildID:   book.ChildID,
		CreatedAt: book.CreatedAt,
	}

	c.JSON(http.StatusCreated, bookResponse)
}

// GetBooksForChild handles getting books for a specific child
func GetBooksForChild(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	childIDParam := c.Param("childId")
	childID, err := strconv.ParseUint(childIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid child ID",
		})
		return
	}

	// Check permission
	hasPermission, err := services.CheckChildPermission(userID, uint(childID), "VIEW")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to check permission: " + err.Error(),
		})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	books, err := services.GetBooksByChild(uint(childID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to get books: " + err.Error(),
		})
		return
	}

	var bookResponses []models.BookResponse
	for _, book := range books {
		bookResponses = append(bookResponses, models.BookResponse{
			ID:        book.ID,
			Title:     book.Title,
			Author:    book.Author,
			DateRead:  book.DateRead,
			ChildID:   book.ChildID,
			CreatedAt: book.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, bookResponses)
}

// GetMyBooksReport handles getting all books for report generation
func GetMyBooksReport(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}
	
	// Get all children for user
	children, err := services.GetChildrenWithPermission(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to get children: " + err.Error(),
		})
		return
	}
	
	var childReports []models.ChildReportResponse
	
	for _, child := range children {
		// Get books for this child
		books, err := services.GetBooksByChild(child.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to get books for child: " + err.Error(),
			})
			return
		}
		
		var bookResponses []models.BookResponse
		for _, book := range books {
			bookResponses = append(bookResponses, models.BookResponse{
				ID:        book.ID,
				Title:     book.Title,
				Author:    book.Author,
				DateRead:  book.DateRead,
				ChildID:   book.ChildID,
				CreatedAt: book.CreatedAt,
			})
		}
		
		childReport := models.ChildReportResponse{
			Child: models.ChildResponse{
				ID:        child.ID,
				Name:      child.Name,
				Grade:     child.Grade,
				OwnerID:   child.OwnerID,
				CreatedAt: child.CreatedAt,
			},
			Books:      bookResponses,
			TotalBooks: len(bookResponses),
		}
		childReports = append(childReports, childReport)
	}
	
	report := models.ReportResponse{
		Children: childReports,
	}
	
	c.JSON(http.StatusOK, report)
}