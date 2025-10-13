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

	// Get optional query parameters for month filtering
	year := c.Query("year")
	month := c.Query("month")
	countOnly := c.Query("count_only") == "true"

	var books []models.Book
	var bookCount int
	
	if year != "" && month != "" {
		// Filter by specific month/year
		yearInt, yearErr := strconv.Atoi(year)
		monthInt, monthErr := strconv.Atoi(month)
		
		if yearErr != nil || monthErr != nil || monthInt < 1 || monthInt > 12 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Message: "Invalid year or month parameter",
			})
			return
		}
		
		if countOnly {
			bookCount, err = services.GetBookCountByChildAndMonth(uint(childID), yearInt, monthInt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Message: "Failed to get book count: " + err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{"count": bookCount})
			return
		} else {
			books, err = services.GetBooksByChildAndMonth(uint(childID), yearInt, monthInt)
		}
	} else {
		// Get all books (existing behavior)
		books, err = services.GetBooksByChild(uint(childID))
	}
	
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
	
	// Get year and month from query parameters
	yearStr := c.Query("year")
	monthStr := c.Query("month")
	
	var books []models.Book
	var err error
	
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
		// Get books for this child - either all time or for specific month
		if yearStr != "" && monthStr != "" {
			year, yearErr := strconv.Atoi(yearStr)
			month, monthErr := strconv.Atoi(monthStr)
			if yearErr != nil || monthErr != nil || month < 1 || month > 12 {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Message: "Invalid year or month parameter",
				})
				return
			}
			books, err = services.GetBooksByChildAndMonth(child.ID, year, month)
		} else {
			books, err = services.GetBooksByChild(child.ID)
		}
		
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
				FirstName: child.FirstName,
				LastName:  child.LastName,
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