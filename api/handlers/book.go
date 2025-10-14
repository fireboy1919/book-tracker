package handlers

import (
	"net/http"
	"strconv"

	"github.com/booktracker/api/middleware"
	"github.com/booktracker/api/models"
	"github.com/booktracker/api/services"
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

	bookResponse := convertBookToResponse(book)

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

		bookResponses := convertBooksToResponses(books)

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

	bookResponses := convertBooksToResponses(books)
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

	bookResponse := convertBookToResponse(book)

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

	bookResponse := convertBookToResponse(updatedBook)
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

	bookResponse := convertBookToResponse(book)

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

	// Check permission using cache
	permCache := middleware.GetPermissionCache(c)
	hasPermission, err := permCache.GetOrCheck(userID, uint(childID), "VIEW")
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

	bookResponses := convertBooksToResponses(books)
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
		
		bookResponses := convertBooksToResponses(books)
		
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

// convertBooksToResponses efficiently converts multiple Book models to BookResponse slice
func convertBooksToResponses(books []models.Book) []models.BookResponse {
	if len(books) == 0 {
		return []models.BookResponse{}
	}
	
	// Pre-allocate slice with exact capacity for efficiency
	responses := make([]models.BookResponse, len(books))
	
	// Batch convert all books
	for i, book := range books {
		responses[i] = models.BookResponse{
			ID:             book.ID,
			DateRead:       book.DateRead,
			ChildID:        book.ChildID,
			LexileLevel:    book.LexileLevel,
			IsPartial:      book.IsPartial,
			PartialComment: book.PartialComment,
			CreatedAt:      book.CreatedAt,
		}
		
		// Set book details based on whether it's a shared book or custom book
		if book.SharedBookID != nil && book.SharedBook != nil {
			// This is a shared book from Open Library
			responses[i].ISBN = book.SharedBook.ISBN
			responses[i].Title = book.SharedBook.Title
			responses[i].Author = book.SharedBook.Author
			responses[i].CoverURL = book.SharedBook.CoverURL
			responses[i].IsCustomBook = false
			responses[i].SharedBookID = book.SharedBookID
			// Lexile level is always from the user's reading record (per-user)
			responses[i].LexileLevel = book.LexileLevel
		} else {
			// This is a custom book
			responses[i].ISBN = book.CustomISBN
			responses[i].Title = book.CustomTitle
			responses[i].Author = book.CustomAuthor
			responses[i].IsCustomBook = true
		}
	}
	
	return responses
}

// convertBookToResponse converts a single Book model to BookResponse (for single book operations)
func convertBookToResponse(book *models.Book) models.BookResponse {
	response := models.BookResponse{
		ID:             book.ID,
		DateRead:       book.DateRead,
		ChildID:        book.ChildID,
		LexileLevel:    book.LexileLevel,
		IsPartial:      book.IsPartial,
		PartialComment: book.PartialComment,
		CreatedAt:      book.CreatedAt,
	}
	
	// Set book details based on whether it's a shared book or custom book
	if book.SharedBookID != nil && book.SharedBook != nil {
		// This is a shared book from Open Library
		response.ISBN = book.SharedBook.ISBN
		response.Title = book.SharedBook.Title
		response.Author = book.SharedBook.Author
		response.CoverURL = book.SharedBook.CoverURL
		response.IsCustomBook = false
		response.SharedBookID = book.SharedBookID
		// Lexile level is always from the user's reading record (per-user)
		response.LexileLevel = book.LexileLevel
	} else {
		// This is a custom book
		response.ISBN = book.CustomISBN
		response.Title = book.CustomTitle
		response.Author = book.CustomAuthor
		response.IsCustomBook = true
	}
	
	return response
}

// CreateCustomBookForChild handles creating a custom book for a specific child
func CreateCustomBookForChild(c *gin.Context) {
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

	var req models.CreateCustomBookRequest
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

	book, err := services.CreateCustomBook(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to create book: " + err.Error(),
		})
		return
	}

	bookResponse := convertBookToResponse(book)
	c.JSON(http.StatusCreated, bookResponse)
}