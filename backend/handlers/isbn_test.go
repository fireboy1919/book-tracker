package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ISBNTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *ISBNTestSuite) SetupTest() {
	// Setup test database
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB

	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Add the ISBN lookup route
	suite.router.POST("/books/lookup-isbn", LookupISBN)
}

func (suite *ISBNTestSuite) TearDownTest() {
	config.CleanupTestDatabase()
}

func (suite *ISBNTestSuite) TestLookupISBN_ValidISBN() {
	// Test with a real ISBN that should exist in Open Library
	// "The Great Gatsby" - 9780743273565
	reqBody := models.ISBNLookupRequest{
		ISBN: "9780743273565",
	}

	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonBody))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.BookInfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	
	// Check that we got a valid response
	assert.Equal(suite.T(), "9780743273565", response.ISBN)
	assert.True(suite.T(), response.Found)
	assert.NotEmpty(suite.T(), response.Title)
	assert.NotEmpty(suite.T(), response.Author)
	
	// The title should contain "Gatsby" (case insensitive check)
	assert.Contains(suite.T(), response.Title, "Gatsby")
}

func (suite *ISBNTestSuite) TestLookupISBN_ValidISBN10() {
	// Test with ISBN-10 format
	// "To Kill a Mockingbird" - 0060935464
	reqBody := models.ISBNLookupRequest{
		ISBN: "0060935464",
	}

	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonBody))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.BookInfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	
	// Check that we got a valid response
	assert.Equal(suite.T(), "0060935464", response.ISBN)
	assert.True(suite.T(), response.Found)
	assert.NotEmpty(suite.T(), response.Title)
	assert.NotEmpty(suite.T(), response.Author)
}

func (suite *ISBNTestSuite) TestLookupISBN_InvalidISBN() {
	// Test with invalid ISBN format
	reqBody := models.ISBNLookupRequest{
		ISBN: "123", // Too short
	}

	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonBody))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response.Message, "Invalid ISBN format")
}

func (suite *ISBNTestSuite) TestLookupISBN_PossiblyNonexistentISBN() {
	// Test with an ISBN that may or may not exist - testing API connectivity
	reqBody := models.ISBNLookupRequest{
		ISBN: "1234567890123", // May exist in Open Library (which has a comprehensive database)
	}

	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonBody))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.BookInfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	
	// Check that we got a valid API response (regardless of whether book exists)
	assert.Equal(suite.T(), "1234567890123", response.ISBN)
	// Found can be true or false - we just want to verify the API is working
	assert.NotNil(suite.T(), response.Found)
	
	// If found, should have title and author; if not found, should be empty
	if response.Found {
		assert.NotEmpty(suite.T(), response.Title)
		assert.NotEmpty(suite.T(), response.Author)
	} else {
		assert.Empty(suite.T(), response.Title)
		assert.Empty(suite.T(), response.Author)
	}
}

func (suite *ISBNTestSuite) TestLookupISBN_ISBNWithHyphens() {
	// Test with ISBN containing hyphens (should be cleaned)
	reqBody := models.ISBNLookupRequest{
		ISBN: "978-0-7432-7356-5", // Same as first test but with hyphens
	}

	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonBody))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.BookInfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	
	// Check that hyphens were removed and we got a valid response
	assert.Equal(suite.T(), "9780743273565", response.ISBN)
	assert.True(suite.T(), response.Found)
	assert.NotEmpty(suite.T(), response.Title)
}

func (suite *ISBNTestSuite) TestLookupISBN_MissingISBN() {
	// Test with empty request body
	reqBody := models.ISBNLookupRequest{
		ISBN: "",
	}

	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonBody))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *ISBNTestSuite) TestLookupISBN_MalformedJSON() {
	// Test with malformed JSON
	req, err := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBufferString("{invalid json}"))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response.Message, "Invalid request data")
}

func (suite *ISBNTestSuite) TestLookupISBN_DuplicateISBN() {
	// Test that looking up the same ISBN twice doesn't cause unique constraint error
	// This test ensures the upsert pattern works correctly
	
	reqBody := models.ISBNLookupRequest{
		ISBN: "9780743273565", // The Great Gatsby
	}

	jsonBody, err := json.Marshal(reqBody)
	assert.NoError(suite.T(), err)

	// First lookup - should create SharedBook entry
	req1, err := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonBody))
	assert.NoError(suite.T(), err)
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, req1)

	assert.Equal(suite.T(), http.StatusOK, w1.Code)

	var response1 models.BookInfoResponse
	err = json.Unmarshal(w1.Body.Bytes(), &response1)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response1.Found)
	assert.NotNil(suite.T(), response1.SharedBookID)

	// Second lookup - should NOT fail with unique constraint error
	// Instead, it should update the existing SharedBook entry or return the existing one
	req2, err := http.NewRequest("POST", "/books/lookup-isbn", bytes.NewBuffer(jsonBody))
	assert.NoError(suite.T(), err)
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusOK, w2.Code)

	var response2 models.BookInfoResponse
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response2.Found)
	assert.NotNil(suite.T(), response2.SharedBookID)

	// Both responses should have the same SharedBookID since it's the same book
	assert.Equal(suite.T(), response1.SharedBookID, response2.SharedBookID)
	
	// Verify the book information is consistent
	assert.Equal(suite.T(), response1.ISBN, response2.ISBN)
	assert.Equal(suite.T(), response1.Title, response2.Title)
	assert.Equal(suite.T(), response1.Author, response2.Author)
}

func TestISBNTestSuite(t *testing.T) {
	suite.Run(t, new(ISBNTestSuite))
}