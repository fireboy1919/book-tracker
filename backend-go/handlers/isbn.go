package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/booktracker/backend-go/models"
	"github.com/booktracker/backend-go/services"
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
	Cover   struct {
		Small  string `json:"small"`
		Medium string `json:"medium"`
		Large  string `json:"large"`
	} `json:"cover"`
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

	// Try original ISBN first
	bookData, finalISBN, found := lookupSingleISBN(isbn)
	if !found {
		c.JSON(http.StatusOK, models.BookInfoResponse{
			ISBN:  isbn,
			Found: false,
		})
		return
	}

	// If no cover image, try to find a related ISBN with better cover
	if bookData.Cover.Small == "" && bookData.Cover.Medium == "" && bookData.Cover.Large == "" {
		betterData, betterISBN, foundBetter := findISBNWithCover(bookData)
		if foundBetter {
			bookData = betterData
			finalISBN = betterISBN
		}
	}

	// Extract author name (take first author if multiple)
	author := ""
	if len(bookData.Authors) > 0 {
		author = bookData.Authors[0].Name
	}

	// Extract cover URL (prefer medium size)
	coverURL := ""
	if bookData.Cover.Medium != "" {
		coverURL = bookData.Cover.Medium
	} else if bookData.Cover.Large != "" {
		coverURL = bookData.Cover.Large
	} else if bookData.Cover.Small != "" {
		coverURL = bookData.Cover.Small
	}

	// Check if this book already exists in SharedBook table
	var existingSharedBook models.SharedBook
	var sharedBookID *uint
	if err := services.GetDB().Where("isbn = ?", isbn).First(&existingSharedBook).Error; err == nil {
		sharedBookID = &existingSharedBook.ID
		// Update cover URL if we have a better one
		if coverURL != "" && existingSharedBook.CoverURL != coverURL {
			services.GetDB().Model(&existingSharedBook).Update("cover_url", coverURL)
		}
	} else {
		// Create new SharedBook entry (no lexile level - that's per-user)
		newSharedBook := models.SharedBook{
			ISBN:     isbn,
			Title:    bookData.Title,
			Author:   author,
			CoverURL: coverURL,
			Source:   "openlibrary",
		}
		if err := services.GetDB().Create(&newSharedBook).Error; err == nil {
			sharedBookID = &newSharedBook.ID
		}
	}

	bookInfo := models.BookInfoResponse{
		ISBN:         finalISBN, // Use the final ISBN (might be different if we found better cover)
		Title:        bookData.Title,
		Author:       author,
		CoverURL:     coverURL,
		Found:        true,
		SharedBookID: sharedBookID,
		// LexileLevel is not available from Open Library API
		// Users will need to fill this manually or get it from Lexile hub
	}

	c.JSON(http.StatusOK, bookInfo)
}

// lookupSingleISBN performs a single ISBN lookup
func lookupSingleISBN(isbn string) (OpenLibraryResponse, string, bool) {
	url := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", isbn)
	
	resp, err := http.Get(url)
	if err != nil {
		return OpenLibraryResponse{}, isbn, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return OpenLibraryResponse{}, isbn, false
	}

	var apiResponse map[string]OpenLibraryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return OpenLibraryResponse{}, isbn, false
	}

	key := fmt.Sprintf("ISBN:%s", isbn)
	bookData, found := apiResponse[key]
	
	if !found || bookData.Title == "" {
		return OpenLibraryResponse{}, isbn, false
	}

	return bookData, isbn, true
}

// findISBNWithCover tries related ISBNs to find one with a cover image
func findISBNWithCover(originalData OpenLibraryResponse) (OpenLibraryResponse, string, bool) {
	// Collect all related ISBNs from the original response
	var relatedISBNs []string
	relatedISBNs = append(relatedISBNs, originalData.ISBN10...)
	relatedISBNs = append(relatedISBNs, originalData.ISBN13...)

	// Try each related ISBN until we find one with a cover
	for _, relatedISBN := range relatedISBNs {
		if relatedISBN == "" {
			continue
		}
		
		bookData, isbn, found := lookupSingleISBN(relatedISBN)
		if !found {
			continue
		}
		
		// Check if this one has a cover image
		if bookData.Cover.Small != "" || bookData.Cover.Medium != "" || bookData.Cover.Large != "" {
			// Found one with cover! Stop here and return it
			return bookData, isbn, true
		}
	}

	// No ISBN with cover found
	return OpenLibraryResponse{}, "", false
}