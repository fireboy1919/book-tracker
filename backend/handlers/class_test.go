package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
)

type ClassHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	testUser    *models.User
	testTeacher *models.User
	testAdmin   *models.User
}

func (suite *ClassHandlerTestSuite) SetupTest() {
	// Setup test database before each test
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Create test users
	suite.testUser = &models.User{
		Email:         "user@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "User",
		IsAdmin:       false,
		IsTeacher:     false,
		EmailVerified: true,
	}
	config.DB.Create(suite.testUser)

	suite.testTeacher = &models.User{
		Email:         "teacher@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "Teacher",
		IsAdmin:       false,
		IsTeacher:     true,
		EmailVerified: true,
	}
	config.DB.Create(suite.testTeacher)

	suite.testAdmin = &models.User{
		Email:         "admin@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "Admin",
		IsAdmin:       true,
		IsTeacher:     false,
		EmailVerified: true,
	}
	config.DB.Create(suite.testAdmin)

	// Setup routes
	suite.router.POST("/classes", suite.mockAuthMiddleware(suite.testTeacher), CreateClass)
	suite.router.GET("/classes", suite.mockAuthMiddleware(suite.testTeacher), GetClasses)
	suite.router.GET("/classes/available", suite.mockAuthMiddleware(suite.testUser), GetAvailableClasses)
	suite.router.GET("/classes/:id", suite.mockAuthMiddleware(suite.testTeacher), GetClass)
	suite.router.PUT("/classes/:id", suite.mockAuthMiddleware(suite.testTeacher), UpdateClass)
	suite.router.POST("/classes/:id/members", suite.mockAuthMiddleware(suite.testTeacher), AddClassMember)
	suite.router.DELETE("/classes/:id/members/:userId", suite.mockAuthMiddleware(suite.testTeacher), RemoveClassMember)
	suite.router.GET("/classes/:id/students", suite.mockAuthMiddleware(suite.testTeacher), GetClassStudents)
	suite.router.POST("/classes/assign-child", suite.mockAuthMiddleware(suite.testUser), AssignChildToClass)
}

func (suite *ClassHandlerTestSuite) TearDownTest() {
	// Cleanup test database after each test
	config.CleanupTestDatabase()
}

func (suite *ClassHandlerTestSuite) mockAuthMiddleware(user *models.User) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("isAdmin", user.IsAdmin)
		c.Set("isTeacher", user.IsTeacher)
		c.Next()
	}
}

func (suite *ClassHandlerTestSuite) TestCreateClass_Success() {
	reqBody := models.CreateClassRequest{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/classes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.ClassResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Class", response.Name)
	assert.Equal(suite.T(), "A test class", response.Description)
	assert.Equal(suite.T(), 10, response.StudentBooksGoal)
	assert.Equal(suite.T(), 5, response.OtherBooksGoal)
}

func (suite *ClassHandlerTestSuite) TestCreateClass_InvalidJSON() {
	req, _ := http.NewRequest("POST", "/classes", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *ClassHandlerTestSuite) TestGetClasses_Success() {
	// Create a test class and add teacher as member
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	membership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&membership)

	req, _ := http.NewRequest("GET", "/classes", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response []models.ClassResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response, 1)
	assert.Equal(suite.T(), "Test Class", response[0].Name)
}

func (suite *ClassHandlerTestSuite) TestGetAvailableClasses_Success() {
	// Create test classes
	testClass1 := models.Class{
		Name:             "Class 1",
		Description:      "First class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass1)

	testClass2 := models.Class{
		Name:             "Class 2",
		Description:      "Second class",
		StudentBooksGoal: 15,
		OtherBooksGoal:   7,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass2)

	req, _ := http.NewRequest("GET", "/classes/available", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response []models.ClassResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response, 2)
}

func (suite *ClassHandlerTestSuite) TestGetClass_Success() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as member
	membership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&membership)

	req, _ := http.NewRequest("GET", "/classes/"+strconv.Itoa(int(testClass.ID)), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.ClassWithMembersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Class", response.Name)
	assert.Len(suite.T(), response.Members, 1)
	assert.Equal(suite.T(), "TEACHER", response.Members[0].Role)
}

func (suite *ClassHandlerTestSuite) TestGetClass_InvalidID() {
	req, _ := http.NewRequest("GET", "/classes/invalid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *ClassHandlerTestSuite) TestAddClassMember_Success() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as member first
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	reqBody := models.AddClassMemberRequest{
		UserID: suite.testUser.ID,
		Role:   "STUDENT",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/classes/"+strconv.Itoa(int(testClass.ID))+"/members", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.ClassMembershipResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testClass.ID, response.ClassID)
	assert.Equal(suite.T(), suite.testUser.ID, response.UserID)
	assert.Equal(suite.T(), "STUDENT", response.Role)
}

func (suite *ClassHandlerTestSuite) TestGetClassStudents_Success() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as member
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	// Add students
	student1 := models.User{
		Email:         "student1@example.com",
		FirstName:     "Alice",
		LastName:      "Smith",
		EmailVerified: true,
	}
	config.DB.Create(&student1)

	studentMembership1 := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  student1.ID,
		Role:    "STUDENT",
	}
	config.DB.Create(&studentMembership1)

	req, _ := http.NewRequest("GET", "/classes/"+strconv.Itoa(int(testClass.ID))+"/students", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response []models.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response, 1)
	assert.Equal(suite.T(), "Alice", response[0].FirstName)
}

func (suite *ClassHandlerTestSuite) TestAssignChildToClass_Success() {
	// Create test child
	testChild := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
		OwnerID:   suite.testUser.ID,
	}
	config.DB.Create(&testChild)

	// Create test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	reqBody := models.AssignChildToClassRequest{
		ChildID: testChild.ID,
		ClassID: testClass.ID,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/classes/assign-child", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.ChildResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testChild.ID, response.ID)
}

func (suite *ClassHandlerTestSuite) TestGetClassTeachers() {
	// Create a test class
	testClass := models.Class{
		Name:             "Test Class",
		Description:      "A test class",
		StudentBooksGoal: 10,
		OtherBooksGoal:   5,
		CreatedByID:      suite.testTeacher.ID,
	}
	config.DB.Create(&testClass)

	// Add teacher as class member
	teacherMembership := models.ClassMembership{
		ClassID: testClass.ID,
		UserID:  suite.testTeacher.ID,
		Role:    "TEACHER",
	}
	config.DB.Create(&teacherMembership)

	// Setup route
	suite.router.GET("/classes/:id/teachers", func(c *gin.Context) {
		c.Set("userID", suite.testTeacher.ID)
		c.Set("isAdmin", false)
		c.Set("isTeacher", true)
		GetClassTeachers(c)
	})

	// Create request
	req, _ := http.NewRequest("GET", "/classes/"+strconv.Itoa(int(testClass.ID))+"/teachers", nil)
	w := httptest.NewRecorder()

	// Perform request
	suite.router.ServeHTTP(w, req)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var teachers []models.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &teachers)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), teachers, 1)
	assert.Equal(suite.T(), suite.testTeacher.ID, teachers[0].ID)
}

func TestClassHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ClassHandlerTestSuite))
}