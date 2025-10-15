package handlers

import (
	"testing"
	"time"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BookHandlerTestSuite struct {
	suite.Suite
}

func (suite *BookHandlerTestSuite) SetupTest() {
	// Setup test database
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB
}

func (suite *BookHandlerTestSuite) TearDownTest() {
	config.CleanupTestDatabase()
}

func (suite *BookHandlerTestSuite) TestConvertBooksToResponsesEmpty() {
	// Test with empty slice
	books := []models.Book{}
	responses := convertBooksToResponses(books)
	
	assert.NotNil(suite.T(), responses)
	assert.Len(suite.T(), responses, 0)
}

func (suite *BookHandlerTestSuite) TestConvertBooksToResponsesSharedBooks() {
	// Create test shared book
	sharedBook := models.SharedBook{
		ISBN:     "9781234567890",
		Title:    "Test Book",
		Author:   "Test Author",
		CoverURL: "http://example.com/cover.jpg",
		Source:   "openlibrary",
	}
	config.DB.Create(&sharedBook)

	// Create test books
	now := time.Now()
	books := []models.Book{
		{
			ID:           1,
			DateRead:     "2024-01-15",
			ChildID:      1,
			SharedBookID: &sharedBook.ID,
			LexileLevel:  "500L",
			IsPartial:    false,
			CreatedAt:    now,
			SharedBook:   &sharedBook,
		},
		{
			ID:           2,
			DateRead:     "2024-01-16",
			ChildID:      1,
			SharedBookID: &sharedBook.ID,
			LexileLevel:  "600L",
			IsPartial:    true,
			PartialComment: "Read chapters 1-3",
			CreatedAt:    now.Add(time.Hour),
			SharedBook:   &sharedBook,
		},
	}

	responses := convertBooksToResponses(books)

	assert.Len(suite.T(), responses, 2)
	
	// Check first book
	assert.Equal(suite.T(), uint(1), responses[0].ID)
	assert.Equal(suite.T(), "2024-01-15", responses[0].DateRead)
	assert.Equal(suite.T(), uint(1), responses[0].ChildID)
	assert.Equal(suite.T(), "9781234567890", responses[0].ISBN)
	assert.Equal(suite.T(), "Test Book", responses[0].Title)
	assert.Equal(suite.T(), "Test Author", responses[0].Author)
	assert.Equal(suite.T(), "http://example.com/cover.jpg", responses[0].CoverURL)
	assert.Equal(suite.T(), "500L", responses[0].LexileLevel)
	assert.False(suite.T(), responses[0].IsCustomBook)
	assert.False(suite.T(), responses[0].IsPartial)
	assert.Equal(suite.T(), &sharedBook.ID, responses[0].SharedBookID)

	// Check second book
	assert.Equal(suite.T(), uint(2), responses[1].ID)
	assert.Equal(suite.T(), "2024-01-16", responses[1].DateRead)
	assert.True(suite.T(), responses[1].IsPartial)
	assert.Equal(suite.T(), "Read chapters 1-3", responses[1].PartialComment)
	assert.Equal(suite.T(), "600L", responses[1].LexileLevel)
}

func (suite *BookHandlerTestSuite) TestConvertBooksToResponsesCustomBooks() {
	// Create test custom books
	now := time.Now()
	books := []models.Book{
		{
			ID:           1,
			DateRead:     "2024-01-15",
			ChildID:      1,
			CustomTitle:  "Custom Book 1",
			CustomAuthor: "Custom Author 1",
			CustomISBN:   "1234567890",
			LexileLevel:  "400L",
			IsPartial:    false,
			CreatedAt:    now,
		},
		{
			ID:           2,
			DateRead:     "2024-01-16",
			ChildID:      1,
			CustomTitle:  "Custom Book 2",
			CustomAuthor: "Custom Author 2",
			CustomISBN:   "",
			LexileLevel:  "500L",
			IsPartial:    true,
			PartialComment: "First half only",
			CreatedAt:    now.Add(time.Hour),
		},
	}

	responses := convertBooksToResponses(books)

	assert.Len(suite.T(), responses, 2)
	
	// Check first book
	assert.Equal(suite.T(), uint(1), responses[0].ID)
	assert.Equal(suite.T(), "Custom Book 1", responses[0].Title)
	assert.Equal(suite.T(), "Custom Author 1", responses[0].Author)
	assert.Equal(suite.T(), "1234567890", responses[0].ISBN)
	assert.True(suite.T(), responses[0].IsCustomBook)
	assert.False(suite.T(), responses[0].IsPartial)
	assert.Nil(suite.T(), responses[0].SharedBookID)

	// Check second book
	assert.Equal(suite.T(), uint(2), responses[1].ID)
	assert.Equal(suite.T(), "Custom Book 2", responses[1].Title)
	assert.Equal(suite.T(), "Custom Author 2", responses[1].Author)
	assert.Equal(suite.T(), "", responses[1].ISBN)
	assert.True(suite.T(), responses[1].IsCustomBook)
	assert.True(suite.T(), responses[1].IsPartial)
	assert.Equal(suite.T(), "First half only", responses[1].PartialComment)
}

func (suite *BookHandlerTestSuite) TestConvertBooksToResponsesPerformance() {
	// Create a shared book for testing
	sharedBook := models.SharedBook{
		ISBN:     "9781234567890",
		Title:    "Performance Test Book",
		Author:   "Test Author",
		CoverURL: "http://example.com/cover.jpg",
		Source:   "openlibrary",
	}
	config.DB.Create(&sharedBook)

	// Create a large slice of books to test performance
	numBooks := 1000
	books := make([]models.Book, numBooks)
	now := time.Now()
	
	for i := 0; i < numBooks; i++ {
		books[i] = models.Book{
			ID:           uint(i + 1),
			DateRead:     "2024-01-15",
			ChildID:      1,
			SharedBookID: &sharedBook.ID,
			LexileLevel:  "500L",
			IsPartial:    false,
			CreatedAt:    now,
			SharedBook:   &sharedBook,
		}
	}

	// Measure conversion time
	start := time.Now()
	responses := convertBooksToResponses(books)
	duration := time.Since(start)

	assert.Len(suite.T(), responses, numBooks)
	
	// Should complete within reasonable time (adjust threshold as needed)
	assert.True(suite.T(), duration < time.Millisecond*100, 
		"Converting %d books took %v, which seems slow", numBooks, duration)
	
	// Verify first and last books are converted correctly
	assert.Equal(suite.T(), uint(1), responses[0].ID)
	assert.Equal(suite.T(), "Performance Test Book", responses[0].Title)
	assert.Equal(suite.T(), uint(numBooks), responses[numBooks-1].ID)
	assert.Equal(suite.T(), "Performance Test Book", responses[numBooks-1].Title)
}

func (suite *BookHandlerTestSuite) TestConvertSingleBookToResponse() {
	// Create a shared book
	sharedBook := models.SharedBook{
		ISBN:     "9781234567890",
		Title:    "Single Test Book",
		Author:   "Test Author",
		CoverURL: "http://example.com/cover.jpg",
		Source:   "openlibrary",
	}
	config.DB.Create(&sharedBook)

	// Create a single book
	book := models.Book{
		ID:           1,
		DateRead:     "2024-01-15",
		ChildID:      1,
		SharedBookID: &sharedBook.ID,
		LexileLevel:  "500L",
		IsPartial:    false,
		CreatedAt:    time.Now(),
		SharedBook:   &sharedBook,
	}

	response := convertBookToResponse(&book)

	assert.Equal(suite.T(), uint(1), response.ID)
	assert.Equal(suite.T(), "9781234567890", response.ISBN)
	assert.Equal(suite.T(), "Single Test Book", response.Title)
	assert.Equal(suite.T(), "Test Author", response.Author)
	assert.Equal(suite.T(), "http://example.com/cover.jpg", response.CoverURL)
	assert.False(suite.T(), response.IsCustomBook)
	assert.Equal(suite.T(), &sharedBook.ID, response.SharedBookID)
}

func (suite *BookHandlerTestSuite) TestConvertBooksToResponsesMemoryEfficiency() {
	// Test that pre-allocation works correctly
	numBooks := 100
	books := make([]models.Book, numBooks)
	now := time.Now()
	
	for i := 0; i < numBooks; i++ {
		books[i] = models.Book{
			ID:           uint(i + 1),
			DateRead:     "2024-01-15",
			ChildID:      1,
			CustomTitle:  "Custom Book",
			CustomAuthor: "Custom Author",
			LexileLevel:  "500L",
			CreatedAt:    now,
		}
	}

	responses := convertBooksToResponses(books)

	// Should have exact capacity (no over-allocation)
	assert.Len(suite.T(), responses, numBooks)
	assert.Equal(suite.T(), numBooks, cap(responses), "Response slice should have exact capacity")
}

func TestBookHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BookHandlerTestSuite))
}