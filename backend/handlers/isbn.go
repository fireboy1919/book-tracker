package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
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

	// Check database first for existing SharedBook
	var existingSharedBook models.SharedBook
	if err := services.GetDB().Where("isbn = ?", isbn).First(&existingSharedBook).Error; err == nil {
		// Found in database! Return immediately without API call
		bookInfo := models.BookInfoResponse{
			ISBN:         existingSharedBook.ISBN,
			Title:        existingSharedBook.Title,
			Author:       existingSharedBook.Author,
			CoverURL:     existingSharedBook.CoverURL,
			Found:        true,
			SharedBookID: &existingSharedBook.ID,
		}
		c.JSON(http.StatusOK, bookInfo)
		return
	}

	// Not in database, try API lookup
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

	// Create or update SharedBook entry (upsert pattern)
	newSharedBook := models.SharedBook{
		ISBN:     isbn,
		Title:    bookData.Title,
		Author:   author,
		CoverURL: coverURL,
		Source:   "openlibrary",
	}
	
	// Try to find existing book by ISBN first
	var existingBook models.SharedBook
	var sharedBookID *uint
	if err := services.GetDB().Where("isbn = ?", isbn).First(&existingBook).Error; err == nil {
		// Book exists, update it with new information
		existingBook.Title = bookData.Title
		existingBook.Author = author
		existingBook.CoverURL = coverURL
		existingBook.Source = "openlibrary"
		
		if err := services.GetDB().Save(&existingBook).Error; err == nil {
			sharedBookID = &existingBook.ID
			newSharedBook = existingBook
		}
	} else {
		// Book doesn't exist, create new one
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

// OpenLibrarySearchResponse represents the search response from Open Library API
type OpenLibrarySearchResponse struct {
	NumFound int `json:"num_found"`
	Start    int `json:"start"`
	Docs     []struct {
		Key               string   `json:"key"`
		Title             string   `json:"title"`
		AuthorName        []string `json:"author_name"`
		FirstPublishYear  int      `json:"first_publish_year"`
		ISBN              []string `json:"isbn"`
		CoverI            int      `json:"cover_i"`
		HasFulltext       bool     `json:"has_fulltext"`
		PublishYear       []int    `json:"publish_year"`
		AuthorKey         []string `json:"author_key"`
		Subject           []string `json:"subject"`
		Publisher         []string `json:"publisher"`
		Language          []string `json:"language"`
	} `json:"docs"`
}

// BookSearchRequest represents a request to search for books
type BookSearchRequest struct {
	Title  string `json:"title,omitempty"`
	Author string `json:"author,omitempty"`
	Query  string `json:"query,omitempty"`
}

// BookSearchResult represents a single book search result
type BookSearchResult struct {
	Title            string   `json:"title"`
	Author           string   `json:"author"`
	AuthorNames      []string `json:"authorNames"`
	FirstPublishYear int      `json:"firstPublishYear"`
	ISBN             string   `json:"isbn,omitempty"`
	CoverURL         string   `json:"coverUrl,omitempty"`
	OpenLibraryKey   string   `json:"openLibraryKey"`
}

// BookSearchResponse represents the response from book search
type BookSearchResponse struct {
	Results   []BookSearchResult `json:"results"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	PerPage   int               `json:"perPage"`
}

// SearchBooks handles searching for books by title and/or author
func SearchBooks(c *gin.Context) {
	var req BookSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate that at least one search parameter is provided
	if req.Title == "" && req.Author == "" && req.Query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "At least one of title, author, or query must be provided",
		})
		return
	}

	// Build search URL
	url := "https://openlibrary.org/search.json?"
	params := []string{}
	
	if req.Query != "" {
		params = append(params, fmt.Sprintf("q=%s", strings.ReplaceAll(req.Query, " ", "+")))
	} else {
		if req.Title != "" {
			params = append(params, fmt.Sprintf("title=%s", strings.ReplaceAll(req.Title, " ", "+")))
		}
		if req.Author != "" {
			params = append(params, fmt.Sprintf("author=%s", strings.ReplaceAll(req.Author, " ", "+")))
		}
	}
	
	// Add pagination and limit
	params = append(params, "limit=20")
	url += strings.Join(params, "&")

	// Make API request
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to search books: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Open Library API returned an error",
		})
		return
	}

	var searchResponse OpenLibrarySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to parse search response: " + err.Error(),
		})
		return
	}

	// Convert to our response format
	results := make([]BookSearchResult, 0, len(searchResponse.Docs))
	for _, doc := range searchResponse.Docs {
		// Get primary author
		author := ""
		if len(doc.AuthorName) > 0 {
			author = doc.AuthorName[0]
		}

		// Get primary ISBN
		isbn := ""
		if len(doc.ISBN) > 0 {
			// Prefer ISBN-13, then ISBN-10
			for _, isbnCandidate := range doc.ISBN {
				cleanISBN := strings.ReplaceAll(strings.ReplaceAll(isbnCandidate, "-", ""), " ", "")
				if len(cleanISBN) == 13 && (strings.HasPrefix(cleanISBN, "978") || strings.HasPrefix(cleanISBN, "979")) {
					isbn = cleanISBN
					break
				} else if len(cleanISBN) == 10 && isbn == "" {
					isbn = cleanISBN
				}
			}
		}

		// Generate cover URL
		coverURL := ""
		if doc.CoverI > 0 {
			coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI)
		}

		result := BookSearchResult{
			Title:            doc.Title,
			Author:           author,
			AuthorNames:      doc.AuthorName,
			FirstPublishYear: doc.FirstPublishYear,
			ISBN:             isbn,
			CoverURL:         coverURL,
			OpenLibraryKey:   doc.Key,
		}

		results = append(results, result)
	}

	response := BookSearchResponse{
		Results: results,
		Total:   searchResponse.NumFound,
		Page:    1,
		PerPage: 20,
	}

	c.JSON(http.StatusOK, response)
}

// CreateBookFromSearchRequest represents a request to create a SharedBook from search result
type CreateBookFromSearchRequest struct {
	Title            string `json:"title" binding:"required"`
	Author           string `json:"author" binding:"required"`
	ISBN             string `json:"isbn,omitempty"`
	CoverURL         string `json:"coverUrl,omitempty"`
	FirstPublishYear int    `json:"firstPublishYear,omitempty"`
	OpenLibraryKey   string `json:"openLibraryKey,omitempty"`
}

// CreateBookFromSearch creates a SharedBook from search result and returns BookInfoResponse
func CreateBookFromSearch(c *gin.Context) {
	var req CreateBookFromSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if this book already exists in our database
	var existingSharedBook models.SharedBook
	var sharedBookID *uint

	// First try to find by ISBN if provided
	if req.ISBN != "" {
		if err := services.GetDB().Where("isbn = ?", req.ISBN).First(&existingSharedBook).Error; err == nil {
			sharedBookID = &existingSharedBook.ID
		}
	}

	// If not found by ISBN, try to find by title and author
	if sharedBookID == nil {
		if err := services.GetDB().Where("title = ? AND author = ?", req.Title, req.Author).First(&existingSharedBook).Error; err == nil {
			sharedBookID = &existingSharedBook.ID
		}
	}

	// If still not found, create a new SharedBook
	if sharedBookID == nil {
		newSharedBook := models.SharedBook{
			ISBN:     req.ISBN,
			Title:    req.Title,
			Author:   req.Author,
			CoverURL: req.CoverURL,
			Source:   "openlibrary",
		}

		// Try one more time by ISBN only (in case we missed it due to different title/author)
		if req.ISBN != "" {
			if err := services.GetDB().Where("isbn = ?", req.ISBN).First(&existingSharedBook).Error; err == nil {
				// Found by ISBN, update it
				existingSharedBook.Title = req.Title
				existingSharedBook.Author = req.Author
				existingSharedBook.CoverURL = req.CoverURL
				existingSharedBook.Source = "openlibrary"
				
				if err := services.GetDB().Save(&existingSharedBook).Error; err != nil {
					c.JSON(http.StatusInternalServerError, models.ErrorResponse{
						Message: "Failed to update shared book: " + err.Error(),
					})
					return
				}
				sharedBookID = &existingSharedBook.ID
			} else {
				// Create new book with ISBN
				sharedBookID, existingSharedBook = createOrFindSharedBook(newSharedBook, req.ISBN, req.Title, req.Author, c)
				if sharedBookID == nil {
					return // Error already handled in createOrFindSharedBook
				}
			}
		} else {
			// No ISBN, check if book exists by title and author to avoid duplicates
			if err := services.GetDB().Where("title = ? AND author = ?", req.Title, req.Author).First(&existingSharedBook).Error; err == nil {
				// Found existing book by title/author
				sharedBookID = &existingSharedBook.ID
			} else {
				// Create new book without ISBN
				sharedBookID, existingSharedBook = createOrFindSharedBook(newSharedBook, req.ISBN, req.Title, req.Author, c)
				if sharedBookID == nil {
					return // Error already handled in createOrFindSharedBook
				}
			}
		}
	}

	// Return in BookInfoResponse format for compatibility with frontend
	bookInfo := models.BookInfoResponse{
		ISBN:         existingSharedBook.ISBN,
		Title:        existingSharedBook.Title,
		Author:       existingSharedBook.Author,
		CoverURL:     existingSharedBook.CoverURL,
		Found:        true,
		SharedBookID: sharedBookID,
	}

	c.JSON(http.StatusOK, bookInfo)
}

// createOrFindSharedBook attempts to create a new SharedBook, handling UNIQUE constraint errors
// Returns the sharedBookID and the book record, or nil if there was an error
func createOrFindSharedBook(newSharedBook models.SharedBook, isbn, title, author string, c *gin.Context) (*uint, models.SharedBook) {
	if err := services.GetDB().Create(&newSharedBook).Error; err != nil {
		// If it's a constraint error, try to find the existing book one more time
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
			var existingSharedBook models.SharedBook
			// Race condition or existing book found - try to find it
			if isbn != "" {
				if findErr := services.GetDB().Where("isbn = ?", isbn).First(&existingSharedBook).Error; findErr == nil {
					return &existingSharedBook.ID, existingSharedBook
				}
			} else {
				if findErr := services.GetDB().Where("title = ? AND author = ?", title, author).First(&existingSharedBook).Error; findErr == nil {
					return &existingSharedBook.ID, existingSharedBook
				}
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Book already exists but could not be retrieved",
			})
			return nil, models.SharedBook{}
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to create shared book: " + err.Error(),
			})
			return nil, models.SharedBook{}
		}
	}
	return &newSharedBook.ID, newSharedBook
}