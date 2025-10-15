package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ISBNOptimizationTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *ISBNOptimizationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *ISBNOptimizationTestSuite) SetupTest() {
	// Setup test database
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB

	// Setup router
	suite.router = gin.New()
	books := suite.router.Group("/books")
	{
		books.POST("/lookup-isbn", LookupISBN)
	}
}

func (suite *ISBNOptimizationTestSuite) TearDownTest() {
	config.CleanupTestDatabase()
}

func (suite *ISBNOptimizationTestSuite) TestISBNLookupDatabaseCache() {
	// Pre-populate database with a shared book
	sharedBook := models.SharedBook{
		ISBN:     "9780061120084",
		Title:    "To Kill a Mockingbird",
		Author:   "Harper Lee",
		CoverURL: "http://example.com/mockingbird.jpg",
		Source:   "openlibrary",
	}
	err := config.DB.Create(&sharedBook).Error
	assert.NoError(suite.T(), err)

	// Test ISBN lookup request
	lookupRequest := models.ISBNLookupRequest{
		ISBN: "978-0-06-112008-4", // Same ISBN with hyphens
	}

	jsonData, err := json.Marshal(lookupRequest)
	assert.NoError(suite.T(), err)

	req, _ := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	
	// Measure response time
	start := time.Now()
	suite.router.ServeHTTP(w, req)
	duration := time.Since(start)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.BookInfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Should find the book in database
	assert.True(suite.T(), response.Found)
	assert.Equal(suite.T(), "9780061120084", response.ISBN)
	assert.Equal(suite.T(), "To Kill a Mockingbird", response.Title)
	assert.Equal(suite.T(), "Harper Lee", response.Author)
	assert.Equal(suite.T(), "http://example.com/mockingbird.jpg", response.CoverURL)
	assert.Equal(suite.T(), &sharedBook.ID, response.SharedBookID)

	// Should be fast (database lookup, not API call)
	assert.True(suite.T(), duration < time.Millisecond*100, 
		"Database lookup took %v, should be much faster than API call", duration)
}

func (suite *ISBNOptimizationTestSuite) TestISBNLookupCachePerformance() {
	// Pre-populate database with a shared book
	sharedBook := models.SharedBook{
		ISBN:     "9780061120084",
		Title:    "To Kill a Mockingbird",
		Author:   "Harper Lee",
		CoverURL: "http://example.com/mockingbird.jpg",
		Source:   "openlibrary",
	}
	config.DB.Create(&sharedBook)

	lookupRequest := models.ISBNLookupRequest{
		ISBN: "9780061120084",
	}

	jsonData, _ := json.Marshal(lookupRequest)

	// First request (should hit database cache)
	req1, _ := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonData))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	
	start1 := time.Now()
	suite.router.ServeHTTP(w1, req1)
	duration1 := time.Since(start1)

	// Second request (should also hit database cache)
	req2, _ := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonData))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	
	start2 := time.Now()
	suite.router.ServeHTTP(w2, req2)
	duration2 := time.Since(start2)

	// Both should be successful
	assert.Equal(suite.T(), http.StatusOK, w1.Code)
	assert.Equal(suite.T(), http.StatusOK, w2.Code)

	// Both should be fast (database lookups)
	assert.True(suite.T(), duration1 < time.Millisecond*100)
	assert.True(suite.T(), duration2 < time.Millisecond*100)

	// Responses should be identical
	var response1, response2 models.BookInfoResponse
	json.Unmarshal(w1.Body.Bytes(), &response1)
	json.Unmarshal(w2.Body.Bytes(), &response2)

	assert.Equal(suite.T(), response1, response2)
}

func (suite *ISBNOptimizationTestSuite) TestISBNLookupNormalization() {
	// Pre-populate database with ISBN without hyphens
	sharedBook := models.SharedBook{
		ISBN:     "9780061120084",
		Title:    "Test Book",
		Author:   "Test Author",
		CoverURL: "http://example.com/test.jpg",
		Source:   "openlibrary",
	}
	config.DB.Create(&sharedBook)

	testCases := []struct {
		name      string
		inputISBN string
		shouldFind bool
	}{
		{
			name:      "ISBN with hyphens",
			inputISBN: "978-0-06-112008-4",
			shouldFind: true,
		},
		{
			name:      "ISBN with spaces",
			inputISBN: "978 0 06 112008 4",
			shouldFind: true,
		},
		{
			name:      "ISBN without formatting",
			inputISBN: "9780061120084",
			shouldFind: true,
		},
		{
			name:      "Mixed formatting",
			inputISBN: "978-0 06-112008 4",
			shouldFind: true,
		},
		{
			name:      "Different ISBN",
			inputISBN: "9781111111111", // Clearly fake ISBN
			shouldFind: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lookupRequest := models.ISBNLookupRequest{
				ISBN: tc.inputISBN,
			}

			jsonData, _ := json.Marshal(lookupRequest)
			req, _ := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var response models.BookInfoResponse
			json.Unmarshal(w.Body.Bytes(), &response)

			if tc.shouldFind {
				assert.True(suite.T(), response.Found, "Should find book for ISBN: %s", tc.inputISBN)
				assert.Equal(suite.T(), "Test Book", response.Title)
				assert.Equal(suite.T(), "Test Author", response.Author)
				assert.NotNil(suite.T(), response.SharedBookID)
			} else {
				assert.False(suite.T(), response.Found, "Should not find book for ISBN: %s", tc.inputISBN)
			}
		})
	}
}

func (suite *ISBNOptimizationTestSuite) TestISBNLookupFallbackToAPI() {
	// Test with ISBN not in database (should fall back to API)
	// Note: This test may be slow as it hits the real OpenLibrary API
	// In a real test suite, you might want to mock the HTTP client

	lookupRequest := models.ISBNLookupRequest{
		ISBN: "9780134685991", // Effective Java by Joshua Bloch
	}

	jsonData, _ := json.Marshal(lookupRequest)
	req, _ := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should succeed (either find the book or gracefully handle not found)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.BookInfoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// If found, should have created a SharedBook entry
	if response.Found {
		assert.NotEmpty(suite.T(), response.Title)
		assert.NotEmpty(suite.T(), response.Author)
		assert.NotNil(suite.T(), response.SharedBookID)

		// Verify it was saved to database
		var sharedBook models.SharedBook
		err := config.DB.Where("isbn = ?", "9780134685991").First(&sharedBook).Error
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), response.Title, sharedBook.Title)
	}
}

func (suite *ISBNOptimizationTestSuite) TestISBNLookupInvalidFormat() {
	testCases := []struct {
		name    string
		isbn    string
		isValid bool
	}{
		{
			name:    "Valid ISBN-13",
			isbn:    "9780061120084",
			isValid: true,
		},
		{
			name:    "Valid ISBN-10",
			isbn:    "0061120081",
			isValid: true,
		},
		{
			name:    "Too short",
			isbn:    "123456789",
			isValid: false,
		},
		{
			name:    "Too long",
			isbn:    "12345678901234",
			isValid: false,
		},
		{
			name:    "Empty string",
			isbn:    "",
			isValid: false,
		},
		{
			name:    "Only hyphens and spaces",
			isbn:    "---   ---",
			isValid: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			lookupRequest := models.ISBNLookupRequest{
				ISBN: tc.isbn,
			}

			jsonData, _ := json.Marshal(lookupRequest)
			req, _ := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			if tc.isValid {
				assert.Equal(suite.T(), http.StatusOK, w.Code)
			} else {
				assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestISBNOptimizationTestSuite(t *testing.T) {
	suite.Run(t, new(ISBNOptimizationTestSuite))
}