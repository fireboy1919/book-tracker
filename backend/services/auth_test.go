package services

import (
	"strings"
	"testing"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	suite.Suite
}

func (suite *AuthServiceTestSuite) SetupTest() {
	// Setup test database before each test
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB
}

func (suite *AuthServiceTestSuite) TearDownTest() {
	// Cleanup test database after each test
	config.CleanupTestDatabase()
}

func (suite *AuthServiceTestSuite) TestHashPassword() {
	// Test that passwords are hashed consistently
	password := "testPassword123"
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)
	assert.NotEqual(suite.T(), password, hash1, "Hash should be different from password")
	assert.NotEqual(suite.T(), password, hash2, "Hash should be different from password")
	assert.NotEqual(suite.T(), hash1, hash2, "BCrypt uses salt, so hashes should be different")
}

func (suite *AuthServiceTestSuite) TestHashPasswordBCryptCompatible() {
	// Test that created hashes are BCrypt compatible
	password := "testPassword123"
	hash, err := HashPassword(password)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), strings.HasPrefix(hash, "$2a$"), "Hash should start with BCrypt prefix")
}

func (suite *AuthServiceTestSuite) TestVerifyPassword() {
	// Test password verification
	password := "testPassword123"
	hash, err := HashPassword(password)
	assert.NoError(suite.T(), err)

	// Correct password should verify
	err = VerifyPassword(password, hash)
	assert.NoError(suite.T(), err)

	// Wrong password should fail
	err = VerifyPassword("wrongPassword", hash)
	assert.Error(suite.T(), err)
}

func (suite *AuthServiceTestSuite) TestAuthenticateUserNonExistent() {
	// Test authentication with non-existent user
	user, err := AuthenticateUser("nonexistent@example.com", "password")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Equal(suite.T(), "invalid credentials", err.Error())
}

func (suite *AuthServiceTestSuite) TestAuthenticateUserWrongPassword() {
	// First create a user
	_, err := CreateUser(models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "correctPassword",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	})
	assert.NoError(suite.T(), err)

	// Try to authenticate with wrong password
	user, err := AuthenticateUser("test@example.com", "wrongPassword")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Equal(suite.T(), "invalid credentials", err.Error())
}

func (suite *AuthServiceTestSuite) TestAuthenticateUserSuccess() {
	// Create a user
	password := "correctPassword"
	_, err := CreateUser(models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  password,
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	})
	assert.NoError(suite.T(), err)

	// Authenticate with correct credentials
	user, err := AuthenticateUser("test@example.com", password)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "test@example.com", user.Email)
	assert.Equal(suite.T(), "Test", user.FirstName)
	assert.Equal(suite.T(), "User", user.LastName)
}

func (suite *AuthServiceTestSuite) TestGenerateToken() {
	// Create a user
	user := &models.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	token, err := GenerateToken(user)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)
	assert.True(suite.T(), len(token) > 50, "Token should be reasonably long")
}

func (suite *AuthServiceTestSuite) TestValidateTokenInvalid() {
	// Test with invalid token
	claims, err := ValidateToken("invalid.token.here")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), claims)
}

func (suite *AuthServiceTestSuite) TestValidateTokenValid() {
	// Create user and generate token
	user := &models.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	token, err := GenerateToken(user)
	assert.NoError(suite.T(), err)

	// Validate the token
	claims, err := ValidateToken(token)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), claims)
	assert.Equal(suite.T(), user.ID, claims.UserID)
	assert.Equal(suite.T(), user.Email, claims.Email)
}

func (suite *AuthServiceTestSuite) TestLoginNonExistentUser() {
	// Test login with non-existent user
	loginReq := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	loginResponse, err := Login(loginReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), loginResponse)
}

func (suite *AuthServiceTestSuite) TestLoginWrongPassword() {
	// Create a user
	_, err := CreateUser(models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "correctPassword",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	})
	assert.NoError(suite.T(), err)

	// Try to login with wrong password
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongPassword",
	}

	loginResponse, err := Login(loginReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), loginResponse)
}

func (suite *AuthServiceTestSuite) TestLoginSuccess() {
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

	// Create a user
	password := "correctPassword"
	_, err := CreateUser(models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  password,
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	})
	assert.NoError(suite.T(), err)

	// Login with correct credentials
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	loginResponse, err := Login(loginReq)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), loginResponse)
	assert.NotEmpty(suite.T(), loginResponse.Token)
	assert.Equal(suite.T(), "test@example.com", loginResponse.User.Email)
	assert.Equal(suite.T(), "Test", loginResponse.User.FirstName)
	assert.Equal(suite.T(), "User", loginResponse.User.LastName)
	assert.False(suite.T(), loginResponse.User.IsAdmin)
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}