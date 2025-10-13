package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
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

func TestISBNTestSuite(t *testing.T) {
	suite.Run(t, new(ISBNTestSuite))
}