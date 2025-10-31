package handlers

import (
	"testing"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BookMonthFilteringIntegrationTestSuite struct {
	suite.Suite
	testUser  *models.User
	testChild *models.Child
}

func (suite *BookMonthFilteringIntegrationTestSuite) SetupTest() {
	// Setup test database
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB

	// Create test user
	userReq := models.CreateUserRequest{
		Email:     "testintegration@example.com",
		Password:  "password123",
		FirstName: "Integration",
		LastName:  "Test",
	}

	user, err := services.CreateUser(userReq)
	assert.NoError(suite.T(), err)
	suite.testUser = user

	// Create test child
	childReq := models.CreateChildRequest{
		FirstName: "Integration",
		LastName:  "Child",
		Grade:     "3rd",
	}

	child, err := services.CreateChild(childReq, suite.testUser.ID)
	assert.NoError(suite.T(), err)
	suite.testChild = child
}

func (suite *BookMonthFilteringIntegrationTestSuite) TearDownTest() {
	config.CleanupTestDatabase()
}

func (suite *BookMonthFilteringIntegrationTestSuite) TestServiceLayerMonthFilteringIntegration() {
	// Create books across different months
	booksData := []struct {
		title    string
		author   string
		dateRead string
	}{
		{"January Book", "Author A", "2023-01-15"},
		{"February Book 1", "Author B", "2023-02-10"},
		{"February Book 2", "Author C", "2023-02-25"},
		{"March Book", "Author D", "2023-03-05"},
	}

	// Create all books using the service layer
	for _, bookData := range booksData {
		req := models.CreateBookRequest{
			Title:        bookData.title,
			Author:       bookData.author,
			DateRead:     bookData.dateRead,
			ChildID:      suite.testChild.ID,
			IsCustomBook: true,
		}

		_, err := services.CreateBook(req)
		assert.NoError(suite.T(), err)
	}

	// Test that month filtering works correctly through service layer
	februaryBooks, err := services.GetBooksByChildAndMonth(suite.testChild.ID, 2023, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), februaryBooks, 2)

	// Test the conversion function used by handlers
	bookResponses := convertBooksToResponses(februaryBooks)
	assert.Len(suite.T(), bookResponses, 2)

	// Verify response structure includes the right fields
	for _, response := range bookResponses {
		assert.NotEmpty(suite.T(), response.Title)
		assert.NotEmpty(suite.T(), response.Author)
		assert.NotEmpty(suite.T(), response.DateRead)
		assert.Equal(suite.T(), suite.testChild.ID, response.ChildID)
		assert.True(suite.T(), response.IsCustomBook)
	}

	// Verify sorting (should be by date_read DESC)
	assert.Equal(suite.T(), "February Book 2", bookResponses[0].Title) // 2023-02-25 comes first
	assert.Equal(suite.T(), "February Book 1", bookResponses[1].Title) // 2023-02-10 comes second
}

func (suite *BookMonthFilteringIntegrationTestSuite) TestConvertBooksToResponsesWithMultipleBooks() {
	// Test the conversion function with different book types
	
	// Create a custom book
	customBookReq := models.CreateBookRequest{
		Title:        "Custom Book",
		Author:       "Custom Author",
		DateRead:     "2023-05-01",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
		LexileLevel:  "500L",
	}

	customBook, err := services.CreateBook(customBookReq)
	assert.NoError(suite.T(), err)

	// For this test, just create another custom book to test the conversion function
	// since we're testing the handler logic, not the shared book functionality
	anotherCustomBookReq := models.CreateBookRequest{
		Title:        "Another Custom Book",
		Author:       "Another Author", 
		DateRead:     "2023-05-02",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
		LexileLevel:  "600L",
	}

	anotherCustomBook, err := services.CreateBook(anotherCustomBookReq)
	assert.NoError(suite.T(), err)

	// Get books and convert
	books := []models.Book{*customBook, *anotherCustomBook}
	responses := convertBooksToResponses(books)

	assert.Len(suite.T(), responses, 2)

	// Find the responses by title
	var firstResponse *models.BookResponse
	var secondResponse *models.BookResponse

	for i := range responses {
		if responses[i].Title == "Custom Book" {
			firstResponse = &responses[i]
		} else if responses[i].Title == "Another Custom Book" {
			secondResponse = &responses[i]
		}
	}

	assert.NotNil(suite.T(), firstResponse)
	assert.NotNil(suite.T(), secondResponse)

	// Verify first book fields
	assert.Equal(suite.T(), "Custom Book", firstResponse.Title)
	assert.Equal(suite.T(), "Custom Author", firstResponse.Author)
	assert.Equal(suite.T(), "500L", firstResponse.LexileLevel)
	assert.True(suite.T(), firstResponse.IsCustomBook)

	// Verify second book fields  
	assert.Equal(suite.T(), "Another Custom Book", secondResponse.Title)
	assert.Equal(suite.T(), "Another Author", secondResponse.Author)
	assert.Equal(suite.T(), "600L", secondResponse.LexileLevel)
	assert.True(suite.T(), secondResponse.IsCustomBook)
}

func (suite *BookMonthFilteringIntegrationTestSuite) TestEmptyBooksConversion() {
	// Test edge case with empty books slice
	emptyBooks := []models.Book{}
	responses := convertBooksToResponses(emptyBooks)
	
	assert.Len(suite.T(), responses, 0)
	assert.NotNil(suite.T(), responses) // Should be empty slice, not nil
}

func TestBookMonthFilteringIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(BookMonthFilteringIntegrationTestSuite))
}