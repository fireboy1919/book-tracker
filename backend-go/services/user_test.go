package services

import (
	"testing"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
	suite.Suite
}

func (suite *UserServiceTestSuite) SetupTest() {
	// Setup test database before each test
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB
}

func (suite *UserServiceTestSuite) TearDownTest() {
	// Cleanup test database after each test
	config.CleanupTestDatabase()
}

func (suite *UserServiceTestSuite) TestCreateUserSuccess() {
	// Create a first user to ensure the test user won't be made admin
	firstUser := models.User{
		Email:         "admin@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Admin",
		LastName:      "User",
		IsAdmin:       true,
		EmailVerified: true,
	}
	config.DB.Create(&firstUser)

	req := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	user, err := CreateUser(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "test@example.com", user.Email)
	assert.Equal(suite.T(), "Test", user.FirstName)
	assert.Equal(suite.T(), "User", user.LastName)
	assert.False(suite.T(), user.IsAdmin)
	assert.NotEmpty(suite.T(), user.PasswordHash)
	assert.NotEqual(suite.T(), "password123", user.PasswordHash, "Password should be hashed")
}

func (suite *UserServiceTestSuite) TestCreateUserDuplicateEmail() {
	req := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	// First creation should succeed
	user1, err1 := CreateUser(req)
	assert.NoError(suite.T(), err1)
	assert.NotNil(suite.T(), user1)

	// Second creation with same email should fail
	user2, err2 := CreateUser(req)
	assert.Error(suite.T(), err2)
	assert.Nil(suite.T(), user2)
	assert.Contains(suite.T(), err2.Error(), "already exists")
}

func (suite *UserServiceTestSuite) TestGetUserByIDSuccess() {
	// Create a user first
	req := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	createdUser, err := CreateUser(req)
	assert.NoError(suite.T(), err)

	// Get user by ID
	user, err := GetUserByID(createdUser.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), createdUser.ID, user.ID)
	assert.Equal(suite.T(), "test@example.com", user.Email)
	assert.Equal(suite.T(), "Test", user.FirstName)
	assert.Equal(suite.T(), "User", user.LastName)
}

func (suite *UserServiceTestSuite) TestGetUserByIDNotFound() {
	user, err := GetUserByID(999)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Equal(suite.T(), "user not found", err.Error())
}

func (suite *UserServiceTestSuite) TestGetUserByEmailSuccess() {
	// Create a user first
	req := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	createdUser, err := CreateUser(req)
	assert.NoError(suite.T(), err)

	// Get user by email
	user, err := GetUserByEmail("test@example.com")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), createdUser.ID, user.ID)
	assert.Equal(suite.T(), "test@example.com", user.Email)
}

func (suite *UserServiceTestSuite) TestGetUserByEmailNotFound() {
	user, err := GetUserByEmail("nonexistent@example.com")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Equal(suite.T(), "user not found", err.Error())
}

func (suite *UserServiceTestSuite) TestGetAllUsers() {
	// Create multiple users
	user1Req := models.CreateUserRequest{
		Email:     "user1@example.com",
		Password:  "password123",
		FirstName: "User",
		LastName:  "One",
		IsAdmin:   false,
	}

	user2Req := models.CreateUserRequest{
		Email:     "user2@example.com",
		Password:  "password123",
		FirstName: "User",
		LastName:  "Two",
		IsAdmin:   false,
	}

	_, err1 := CreateUser(user1Req)
	_, err2 := CreateUser(user2Req)
	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)

	// Get all users
	users, err := GetAllUsers()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 2)

	// Check if both users are present
	emails := make([]string, len(users))
	for i, user := range users {
		emails[i] = user.Email
	}
	assert.Contains(suite.T(), emails, "user1@example.com")
	assert.Contains(suite.T(), emails, "user2@example.com")
}

func (suite *UserServiceTestSuite) TestUpdateUserSuccess() {
	// Create a user first
	createReq := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	createdUser, err := CreateUser(createReq)
	assert.NoError(suite.T(), err)

	// Update the user
	updateReq := models.UpdateUserRequest{
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "User",
		IsAdmin:   true,
	}

	updatedUser, err := UpdateUser(createdUser.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedUser)
	assert.Equal(suite.T(), "updated@example.com", updatedUser.Email)
	assert.Equal(suite.T(), "Updated", updatedUser.FirstName)
	assert.Equal(suite.T(), "User", updatedUser.LastName)
	assert.True(suite.T(), updatedUser.IsAdmin)
}

func (suite *UserServiceTestSuite) TestUpdateUserNotFound() {
	updateReq := models.UpdateUserRequest{
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "User",
		IsAdmin:   false,
	}

	updatedUser, err := UpdateUser(999, updateReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), updatedUser)
	assert.Equal(suite.T(), "user not found", err.Error())
}

func (suite *UserServiceTestSuite) TestUpdateUserDuplicateEmail() {
	// Create two users
	user1Req := models.CreateUserRequest{
		Email:     "user1@example.com",
		Password:  "password123",
		FirstName: "User",
		LastName:  "One",
		IsAdmin:   false,
	}

	user2Req := models.CreateUserRequest{
		Email:     "user2@example.com",
		Password:  "password123",
		FirstName: "User",
		LastName:  "Two",
		IsAdmin:   false,
	}

	_, err1 := CreateUser(user1Req)
	user2, err2 := CreateUser(user2Req)
	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)

	// Try to update user2 with user1's email
	updateReq := models.UpdateUserRequest{
		Email:     "user1@example.com", // This email is already taken
		FirstName: "Updated",
		LastName:  "User",
		IsAdmin:   false,
	}

	updatedUser, err := UpdateUser(user2.ID, updateReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), updatedUser)
	assert.Contains(suite.T(), err.Error(), "already taken")
}

func (suite *UserServiceTestSuite) TestDeleteUserSuccess() {
	// Create a user first
	req := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	createdUser, err := CreateUser(req)
	assert.NoError(suite.T(), err)

	// Delete the user
	err = DeleteUser(createdUser.ID)
	assert.NoError(suite.T(), err)

	// Verify user is deleted
	user, err := GetUserByID(createdUser.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
}

func (suite *UserServiceTestSuite) TestDeleteUserNotFound() {
	err := DeleteUser(999)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "user not found", err.Error())
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}