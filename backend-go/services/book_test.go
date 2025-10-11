package services

import (
	"testing"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
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
		Name: "Test Child",
		Age:  8,
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
		Title:    "Test Book",
		Author:   "Test Author",
		DateRead: "2023-10-01",
		ChildID:  suite.testChild.ID,
	}

	book, err := CreateBook(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), book)
	assert.Equal(suite.T(), "Test Book", book.Title)
	assert.Equal(suite.T(), "Test Author", book.Author)
	assert.Equal(suite.T(), "2023-10-01", book.DateRead)
	assert.Equal(suite.T(), suite.testChild.ID, book.ChildID)
}

func (suite *BookServiceTestSuite) TestGetBookByIDSuccess() {
	// Create a book first
	req := models.CreateBookRequest{
		Title:    "Test Book",
		Author:   "Test Author",
		DateRead: "2023-10-01",
		ChildID:  suite.testChild.ID,
	}

	createdBook, err := CreateBook(req)
	assert.NoError(suite.T(), err)

	// Get book by ID
	book, err := GetBookByID(createdBook.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), book)
	assert.Equal(suite.T(), createdBook.ID, book.ID)
	assert.Equal(suite.T(), "Test Book", book.Title)
	assert.Equal(suite.T(), "Test Author", book.Author)
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
		Title:    "Book One",
		Author:   "Author One",
		DateRead: "2023-10-01",
		ChildID:  suite.testChild.ID,
	}

	book2Req := models.CreateBookRequest{
		Title:    "Book Two",
		Author:   "Author Two",
		DateRead: "2023-10-02",
		ChildID:  suite.testChild.ID,
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
	assert.Equal(suite.T(), "Book Two", books[0].Title) // More recent date should come first
	assert.Equal(suite.T(), "Book One", books[1].Title)
}

func (suite *BookServiceTestSuite) TestGetBooksForUser() {
	// Create a book
	bookReq := models.CreateBookRequest{
		Title:    "Test Book",
		Author:   "Test Author",
		DateRead: "2023-10-01",
		ChildID:  suite.testChild.ID,
	}

	_, err := CreateBook(bookReq)
	assert.NoError(suite.T(), err)

	// Get books for user (owner should see their child's books)
	books, err := GetBooksForUser(suite.testUser.ID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), books, 1)
	assert.Equal(suite.T(), "Test Book", books[0].Title)
}

func (suite *BookServiceTestSuite) TestUpdateBookSuccess() {
	// Create a book first
	createReq := models.CreateBookRequest{
		Title:    "Original Title",
		Author:   "Original Author",
		DateRead: "2023-10-01",
		ChildID:  suite.testChild.ID,
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
	assert.Equal(suite.T(), "Updated Title", updatedBook.Title)
	assert.Equal(suite.T(), "Updated Author", updatedBook.Author)
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
		Title:    "Test Book",
		Author:   "Test Author",
		DateRead: "2023-10-01",
		ChildID:  suite.testChild.ID,
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
		Name: "Second Child",
		Age:  10,
	}

	child2, err := CreateChild(child2Req, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Create books for both children
	book1Req := models.CreateBookRequest{
		Title:    "Book for Child 1",
		Author:   "Author One",
		DateRead: "2023-10-01",
		ChildID:  suite.testChild.ID,
	}

	book2Req := models.CreateBookRequest{
		Title:    "Book for Child 2",
		Author:   "Author Two",
		DateRead: "2023-10-02",
		ChildID:  child2.ID,
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
		titles[i] = book.Title
	}
	assert.Contains(suite.T(), titles, "Book for Child 1")
	assert.Contains(suite.T(), titles, "Book for Child 2")
}

func TestBookServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BookServiceTestSuite))
}