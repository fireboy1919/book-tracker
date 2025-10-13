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

type AuthHandlerTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *AuthHandlerTestSuite) SetupSuite() {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	// Setup test database before each test
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB

	// Setup router
	suite.router = gin.New()
	auth := suite.router.Group("/auth")
	{
		auth.POST("/register", RegisterUser)
		auth.POST("/login", LoginUser)
	}
}

func (suite *AuthHandlerTestSuite) TearDownTest() {
	// Cleanup test database after each test
	config.CleanupTestDatabase()
}

func (suite *AuthHandlerTestSuite) TestRegisterUserSuccess() {
	createUserRequest := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	jsonData, err := json.Marshal(createUserRequest)
	assert.NoError(suite.T(), err)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test@example.com", response.Email)
	assert.Equal(suite.T(), "Test", response.FirstName)
	assert.Equal(suite.T(), "User", response.LastName)
	assert.False(suite.T(), response.IsAdmin)
}

func (suite *AuthHandlerTestSuite) TestRegisterUserDuplicateEmail() {
	createUserRequest := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	jsonData, err := json.Marshal(createUserRequest)
	assert.NoError(suite.T(), err)

	// First registration should succeed
	req1, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, req1)
	assert.Equal(suite.T(), http.StatusCreated, w1.Code)

	// Second registration with same email should fail
	req2, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)
	assert.Equal(suite.T(), http.StatusBadRequest, w2.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w2.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse.Message, "already exists")
}

func (suite *AuthHandlerTestSuite) TestRegisterUserInvalidRequest() {
	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *AuthHandlerTestSuite) TestRegisterUserMissingFields() {
	// Test with missing required fields
	incompleteRequest := map[string]string{
		"email": "test@example.com",
		// Missing password, firstName, lastName
	}

	jsonData, err := json.Marshal(incompleteRequest)
	assert.NoError(suite.T(), err)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *AuthHandlerTestSuite) TestLoginUserSuccess() {
	// First register a user
	createUserRequest := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	jsonData, err := json.Marshal(createUserRequest)
	assert.NoError(suite.T(), err)

	regReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	regReq.Header.Set("Content-Type", "application/json")

	regW := httptest.NewRecorder()
	suite.router.ServeHTTP(regW, regReq)
	assert.Equal(suite.T(), http.StatusCreated, regW.Code)

	// Then try to login
	loginRequest := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	loginData, err := json.Marshal(loginRequest)
	assert.NoError(suite.T(), err)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var loginResponse models.LoginResponse
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), loginResponse.Token)
	assert.Equal(suite.T(), "test@example.com", loginResponse.User.Email)
	assert.Equal(suite.T(), "Test", loginResponse.User.FirstName)
	assert.Equal(suite.T(), "User", loginResponse.User.LastName)
}

func (suite *AuthHandlerTestSuite) TestLoginUserWrongPassword() {
	// First register a user
	createUserRequest := models.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	jsonData, err := json.Marshal(createUserRequest)
	assert.NoError(suite.T(), err)

	regReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
	regReq.Header.Set("Content-Type", "application/json")

	regW := httptest.NewRecorder()
	suite.router.ServeHTTP(regW, regReq)
	assert.Equal(suite.T(), http.StatusCreated, regW.Code)

	// Try to login with wrong password
	loginRequest := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	loginData, err := json.Marshal(loginRequest)
	assert.NoError(suite.T(), err)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid credentials", errorResponse.Message)
}

func (suite *AuthHandlerTestSuite) TestLoginUserNonExistent() {
	loginRequest := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	jsonData, err := json.Marshal(loginRequest)
	assert.NoError(suite.T(), err)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid credentials", errorResponse.Message)
}

func (suite *AuthHandlerTestSuite) TestLoginUserInvalidRequest() {
	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}