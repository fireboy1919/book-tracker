package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/booktracker/backend-go/models"
	"github.com/gin-gonic/gin"
)

// OpenLibraryResponse represents the response from Open Library API
type OpenLibraryResponse struct {
	Title   string   `json:"title"`
	Authors []struct {
		Name string `json:"name"`
	} `json:"authors"`
	ISBN10  []string `json:"isbn_10"`
	ISBN13  []string `json:"isbn_13"`
}

// LookupISBN handles looking up book information by ISBN
func LookupISBN(c *gin.Context) {
	var req models.ISBNLookupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Clean ISBN (remove hyphens, spaces)
	isbn := strings.ReplaceAll(strings.ReplaceAll(req.ISBN, "-", ""), " ", "")
	
	// Validate ISBN format (basic check)
	if len(isbn) != 10 && len(isbn) != 13 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid ISBN format. Must be 10 or 13 digits.",
		})
		return
	}

	// Call Open Library API
	url := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", isbn)
	
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to lookup ISBN: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, models.BookInfoResponse{
			ISBN:  isbn,
			Found: false,
		})
		return
	}

	// Parse response
	var apiResponse map[string]OpenLibraryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to parse API response: " + err.Error(),
		})
		return
	}

	// Check if book was found
	key := fmt.Sprintf("ISBN:%s", isbn)
	bookData, found := apiResponse[key]
	
	if !found || bookData.Title == "" {
		c.JSON(http.StatusOK, models.BookInfoResponse{
			ISBN:  isbn,
			Found: false,
		})
		return
	}

	// Extract author name (take first author if multiple)
	author := ""
	if len(bookData.Authors) > 0 {
		author = bookData.Authors[0].Name
	}

	bookInfo := models.BookInfoResponse{
		ISBN:   isbn,
		Title:  bookData.Title,
		Author: author,
		Found:  true,
		// LexileLevel is not available from Open Library API
		// Users will need to fill this manually or get it from Lexile hub
	}

	c.JSON(http.StatusOK, bookInfo)
}