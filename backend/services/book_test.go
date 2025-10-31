package services

import (
	"testing"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BookServiceTestSuite struct {
	suite.Suite
	testUser  *models.User
	testChild *models.Child
}

func (suite *BookServiceTestSuite) SetupTest() {
	// Setup test database before each test
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB

	// Create a test user for book operations
	userReq := models.CreateUserRequest{
		Email:     "testuser@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	user, err := CreateUser(userReq)
	assert.NoError(suite.T(), err)
	suite.testUser = user

	// Create a test child for book operations
	childReq := models.CreateChildRequest{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
	}

	child, err := CreateChild(childReq, suite.testUser.ID)
	assert.NoError(suite.T(), err)
	suite.testChild = child
}

func (suite *BookServiceTestSuite) TearDownTest() {
	// Cleanup test database after each test
	config.CleanupTestDatabase()
}

func (suite *BookServiceTestSuite) TestCreateBookSuccess() {
	req := models.CreateBookRequest{
		Title:        "Test Book",
		Author:       "Test Author",
		DateRead:     "2023-10-01",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
	}

	book, err := CreateBook(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), book)
	assert.Equal(suite.T(), "Test Book", book.CustomTitle)
	assert.Equal(suite.T(), "Test Author", book.CustomAuthor)
	assert.Equal(suite.T(), "2023-10-01", book.DateRead)
	assert.Equal(suite.T(), suite.testChild.ID, book.ChildID)
}

func (suite *BookServiceTestSuite) TestGetBookByIDSuccess() {
	// Create a book first
	req := models.CreateBookRequest{
		Title:        "Test Book",
		Author:       "Test Author",
		DateRead:     "2023-10-01",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
	}

	createdBook, err := CreateBook(req)
	assert.NoError(suite.T(), err)

	// Get book by ID
	book, err := GetBookByID(createdBook.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), book)
	assert.Equal(suite.T(), createdBook.ID, book.ID)
	assert.Equal(suite.T(), "Test Book", book.CustomTitle)
	assert.Equal(suite.T(), "Test Author", book.CustomAuthor)
	assert.Equal(suite.T(), "2023-10-01", book.DateRead)
	assert.Equal(suite.T(), suite.testChild.ID, book.ChildID)
}

func (suite *BookServiceTestSuite) TestGetBookByIDNotFound() {
	book, err := GetBookByID(999)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), book)
	assert.Equal(suite.T(), "book not found", err.Error())
}

func (suite *BookServiceTestSuite) TestGetBooksByChild() {
	// Create multiple books for the same child
	book1Req := models.CreateBookRequest{
		Title:        "Book One",
		Author:       "Author One",
		DateRead:     "2023-10-01",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
	}

	book2Req := models.CreateBookRequest{
		Title:        "Book Two",
		Author:       "Author Two",
		DateRead:     "2023-10-02",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
	}

	_, err1 := CreateBook(book1Req)
	_, err2 := CreateBook(book2Req)
	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)

	// Get books by child
	books, err := GetBooksByChild(suite.testChild.ID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), books, 2)

	// Check if both books are present (should be ordered by date_read DESC)
	assert.Equal(suite.T(), "Book Two", books[0].CustomTitle) // More recent date should come first
	assert.Equal(suite.T(), "Book One", books[1].CustomTitle)
}

func (suite *BookServiceTestSuite) TestGetBooksForUser() {
	// Create a book
	bookReq := models.CreateBookRequest{
		Title:        "Test Book",
		Author:       "Test Author",
		DateRead:     "2023-10-01",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
	}

	_, err := CreateBook(bookReq)
	assert.NoError(suite.T(), err)

	// Get books for user (owner should see their child's books)
	books, err := GetBooksForUser(suite.testUser.ID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), books, 1)
	assert.Equal(suite.T(), "Test Book", books[0].CustomTitle)
}

func (suite *BookServiceTestSuite) TestUpdateBookSuccess() {
	// Create a book first
	createReq := models.CreateBookRequest{
		Title:        "Original Title",
		Author:       "Original Author",
		DateRead:     "2023-10-01",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
	}

	createdBook, err := CreateBook(createReq)
	assert.NoError(suite.T(), err)

	// Update the book
	updateReq := models.UpdateBookRequest{
		Title:    "Updated Title",
		Author:   "Updated Author",
		DateRead: "2023-10-02",
	}

	updatedBook, err := UpdateBook(createdBook.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedBook)
	assert.Equal(suite.T(), "Updated Title", updatedBook.CustomTitle)
	assert.Equal(suite.T(), "Updated Author", updatedBook.CustomAuthor)
	assert.Equal(suite.T(), "2023-10-02", updatedBook.DateRead)
	assert.Equal(suite.T(), createdBook.ID, updatedBook.ID)
	assert.Equal(suite.T(), suite.testChild.ID, updatedBook.ChildID) // ChildID should remain unchanged
}

func (suite *BookServiceTestSuite) TestUpdateBookNotFound() {
	updateReq := models.UpdateBookRequest{
		Title:    "Updated Title",
		Author:   "Updated Author",
		DateRead: "2023-10-02",
	}

	updatedBook, err := UpdateBook(999, updateReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), updatedBook)
	assert.Equal(suite.T(), "book not found", err.Error())
}

func (suite *BookServiceTestSuite) TestDeleteBookSuccess() {
	// Create a book first
	req := models.CreateBookRequest{
		Title:        "Test Book",
		Author:       "Test Author",
		DateRead:     "2023-10-01",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
	}

	createdBook, err := CreateBook(req)
	assert.NoError(suite.T(), err)

	// Delete the book
	err = DeleteBook(createdBook.ID)
	assert.NoError(suite.T(), err)

	// Verify book is deleted
	book, err := GetBookByID(createdBook.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), book)
}

func (suite *BookServiceTestSuite) TestDeleteBookNotFound() {
	err := DeleteBook(999)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "book not found", err.Error())
}

func (suite *BookServiceTestSuite) TestGetBooksForUserWithMultipleChildren() {
	// Create another child
	child2Req := models.CreateChildRequest{
		FirstName: "Second",
		LastName:  "Child",
		Grade:     "5th",
	}

	child2, err := CreateChild(child2Req, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Create books for both children
	book1Req := models.CreateBookRequest{
		Title:        "Book for Child 1",
		Author:       "Author One",
		DateRead:     "2023-10-01",
		ChildID:      suite.testChild.ID,
		IsCustomBook: true,
	}

	book2Req := models.CreateBookRequest{
		Title:        "Book for Child 2",
		Author:       "Author Two",
		DateRead:     "2023-10-02",
		ChildID:      child2.ID,
		IsCustomBook: true,
	}

	_, err1 := CreateBook(book1Req)
	_, err2 := CreateBook(book2Req)
	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)

	// Get books for user (should see books from both children)
	books, err := GetBooksForUser(suite.testUser.ID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), books, 2)

	// Check if both books are present
	titles := make([]string, len(books))
	for i, book := range books {
		titles[i] = book.CustomTitle
	}
	assert.Contains(suite.T(), titles, "Book for Child 1")
	assert.Contains(suite.T(), titles, "Book for Child 2")
}

func (suite *BookServiceTestSuite) TestGetBooksByChildAndMonth() {
	// Create books with different read dates
	booksData := []struct {
		title    string
		author   string
		dateRead string
	}{
		{"January Book", "Author A", "2023-01-15"},
		{"February Book 1", "Author B", "2023-02-10"},
		{"February Book 2", "Author C", "2023-02-25"},
		{"March Book", "Author D", "2023-03-05"},
		{"December Book", "Author E", "2023-12-20"},
	}

	// Create all books
	createdBooks := make([]*models.Book, len(booksData))
	for i, bookData := range booksData {
		req := models.CreateBookRequest{
			Title:        bookData.title,
			Author:       bookData.author,
			DateRead:     bookData.dateRead,
			ChildID:      suite.testChild.ID,
			IsCustomBook: true,
		}

		book, err := CreateBook(req)
		assert.NoError(suite.T(), err)
		createdBooks[i] = book
	}

	// Test February 2023 - should return 2 books
	februaryBooks, err := GetBooksByChildAndMonth(suite.testChild.ID, 2023, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), februaryBooks, 2)
	
	// Verify the books are sorted by date_read DESC (newest first)
	assert.Equal(suite.T(), "February Book 2", februaryBooks[0].CustomTitle)
	assert.Equal(suite.T(), "February Book 1", februaryBooks[1].CustomTitle)

	// Test January 2023 - should return 1 book
	januaryBooks, err := GetBooksByChildAndMonth(suite.testChild.ID, 2023, 1)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), januaryBooks, 1)
	assert.Equal(suite.T(), "January Book", januaryBooks[0].CustomTitle)

	// Test March 2023 - should return 1 book
	marchBooks, err := GetBooksByChildAndMonth(suite.testChild.ID, 2023, 3)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), marchBooks, 1)
	assert.Equal(suite.T(), "March Book", marchBooks[0].CustomTitle)

	// Test April 2023 - should return 0 books
	aprilBooks, err := GetBooksByChildAndMonth(suite.testChild.ID, 2023, 4)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), aprilBooks, 0)

	// Test December 2023 - should return 1 book
	decemberBooks, err := GetBooksByChildAndMonth(suite.testChild.ID, 2023, 12)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), decemberBooks, 1)
	assert.Equal(suite.T(), "December Book", decemberBooks[0].CustomTitle)

	// Test different year - should return 0 books
	books2024, err := GetBooksByChildAndMonth(suite.testChild.ID, 2024, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), books2024, 0)

	// Test with different child - should return 0 books
	child2Req := models.CreateChildRequest{
		FirstName: "Other",
		LastName:  "Child",
		Grade:     "4th",
	}
	child2, err := CreateChild(child2Req, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	child2Books, err := GetBooksByChildAndMonth(child2.ID, 2023, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), child2Books, 0)
}

func (suite *BookServiceTestSuite) TestGetBookCountByChildAndMonth() {
	// Create books with different read dates
	booksData := []struct {
		title    string
		dateRead string
	}{
		{"Book 1", "2023-05-01"},
		{"Book 2", "2023-05-15"},
		{"Book 3", "2023-05-31"},
		{"Book 4", "2023-06-01"},
	}

	// Create all books
	for _, bookData := range booksData {
		req := models.CreateBookRequest{
			Title:        bookData.title,
			Author:       "Test Author",
			DateRead:     bookData.dateRead,
			ChildID:      suite.testChild.ID,
			IsCustomBook: true,
		}

		_, err := CreateBook(req)
		assert.NoError(suite.T(), err)
	}

	// Test May 2023 - should return 3
	mayCount, err := GetBookCountByChildAndMonth(suite.testChild.ID, 2023, 5)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, mayCount)

	// Test June 2023 - should return 1
	juneCount, err := GetBookCountByChildAndMonth(suite.testChild.ID, 2023, 6)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, juneCount)

	// Test July 2023 - should return 0
	julyCount, err := GetBookCountByChildAndMonth(suite.testChild.ID, 2023, 7)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, julyCount)

	// Test different year - should return 0
	may2024Count, err := GetBookCountByChildAndMonth(suite.testChild.ID, 2024, 5)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, may2024Count)
}

func (suite *BookServiceTestSuite) TestMonthYearBoundaries() {
	// Test edge cases around month/year boundaries
	booksData := []struct {
		title    string
		dateRead string
	}{
		{"End of February", "2023-02-28"},    // Last day of February
		{"Start of March", "2023-03-01"},     // First day of March
		{"End of Year", "2023-12-31"},        // Last day of year
		{"Start of Next Year", "2024-01-01"}, // First day of next year
	}

	// Create all books
	for _, bookData := range booksData {
		req := models.CreateBookRequest{
			Title:        bookData.title,
			Author:       "Test Author",
			DateRead:     bookData.dateRead,
			ChildID:      suite.testChild.ID,
			IsCustomBook: true,
		}

		_, err := CreateBook(req)
		assert.NoError(suite.T(), err)
	}

	// Test February 2023 - should only include February book
	febBooks, err := GetBooksByChildAndMonth(suite.testChild.ID, 2023, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), febBooks, 1)
	assert.Equal(suite.T(), "End of February", febBooks[0].CustomTitle)

	// Test March 2023 - should only include March book
	marBooks, err := GetBooksByChildAndMonth(suite.testChild.ID, 2023, 3)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), marBooks, 1)
	assert.Equal(suite.T(), "Start of March", marBooks[0].CustomTitle)

	// Test December 2023 - should only include December book
	decBooks, err := GetBooksByChildAndMonth(suite.testChild.ID, 2023, 12)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), decBooks, 1)
	assert.Equal(suite.T(), "End of Year", decBooks[0].CustomTitle)

	// Test January 2024 - should only include January book
	janBooks, err := GetBooksByChildAndMonth(suite.testChild.ID, 2024, 1)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), janBooks, 1)
	assert.Equal(suite.T(), "Start of Next Year", janBooks[0].CustomTitle)
}

func TestBookServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BookServiceTestSuite))
}