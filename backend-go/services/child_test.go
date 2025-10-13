package services

import (
	"testing"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChildServiceTestSuite struct {
	suite.Suite
	testUser *models.User
}

func (suite *ChildServiceTestSuite) SetupTest() {
	// Setup test database before each test
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB

	// Create a test user for child operations
	userReq := models.CreateUserRequest{
		Email:     "testuser@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	user, err := CreateUser(userReq)
	assert.NoError(suite.T(), err)
	suite.testUser = user
}

func (suite *ChildServiceTestSuite) TearDownTest() {
	// Cleanup test database after each test
	config.CleanupTestDatabase()
}

func (suite *ChildServiceTestSuite) TestCreateChildSuccess() {
	req := models.CreateChildRequest{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
	}

	child, err := CreateChild(req, suite.testUser.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), child)
	assert.Equal(suite.T(), "Test", child.FirstName)
	assert.Equal(suite.T(), "Child", child.LastName)
	assert.Equal(suite.T(), "3rd", child.Grade)
	assert.Equal(suite.T(), suite.testUser.ID, child.OwnerID)
}

func (suite *ChildServiceTestSuite) TestGetChildByIDSuccess() {
	// Create a child first
	req := models.CreateChildRequest{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
	}

	createdChild, err := CreateChild(req, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Get child by ID
	child, err := GetChildByID(createdChild.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), child)
	assert.Equal(suite.T(), createdChild.ID, child.ID)
	assert.Equal(suite.T(), "Test", child.FirstName)
	assert.Equal(suite.T(), "Child", child.LastName)
	assert.Equal(suite.T(), "3rd", child.Grade)
	assert.Equal(suite.T(), suite.testUser.ID, child.OwnerID)
}

func (suite *ChildServiceTestSuite) TestGetChildByIDNotFound() {
	child, err := GetChildByID(999)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), child)
	assert.Equal(suite.T(), "child not found", err.Error())
}

func (suite *ChildServiceTestSuite) TestGetChildrenByOwner() {
	// Create multiple children
	child1Req := models.CreateChildRequest{
		FirstName: "Child",
		LastName:  "One",
		Grade:     "2nd",
	}

	child2Req := models.CreateChildRequest{
		FirstName: "Child",
		LastName:  "Two",
		Grade:     "4th",
	}

	_, err1 := CreateChild(child1Req, suite.testUser.ID)
	_, err2 := CreateChild(child2Req, suite.testUser.ID)
	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)

	// Get children by owner
	children, err := GetChildrenByOwner(suite.testUser.ID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), children, 2)

	// Check if both children are present
	lastNames := make([]string, len(children))
	for i, child := range children {
		lastNames[i] = child.LastName
	}
	assert.Contains(suite.T(), lastNames, "One")
	assert.Contains(suite.T(), lastNames, "Two")
}

func (suite *ChildServiceTestSuite) TestGetChildrenWithPermission() {
	// Create a child
	childReq := models.CreateChildRequest{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
	}

	_, err := CreateChild(childReq, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Get children with permission (owner should have access)
	children, err := GetChildrenWithPermission(suite.testUser.ID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), children, 1)
	assert.Equal(suite.T(), "Test", children[0].FirstName)
	assert.Equal(suite.T(), "Child", children[0].LastName)
}

func (suite *ChildServiceTestSuite) TestUpdateChildSuccess() {
	// Create a child first
	createReq := models.CreateChildRequest{
		FirstName: "Original",
		LastName:  "Name",
		Grade:     "2nd",
	}

	createdChild, err := CreateChild(createReq, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Update the child
	updateReq := models.UpdateChildRequest{
		FirstName: "Updated",
		LastName:  "Name",
		Grade:     "3rd",
	}

	updatedChild, err := UpdateChild(createdChild.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedChild)
	assert.Equal(suite.T(), "Updated", updatedChild.FirstName)
	assert.Equal(suite.T(), "Name", updatedChild.LastName)
	assert.Equal(suite.T(), "3rd", updatedChild.Grade)
	assert.Equal(suite.T(), createdChild.ID, updatedChild.ID)
}

func (suite *ChildServiceTestSuite) TestUpdateChildNotFound() {
	updateReq := models.UpdateChildRequest{
		FirstName: "Updated",
		LastName:  "Name",
		Grade:     "3rd",
	}

	updatedChild, err := UpdateChild(999, updateReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), updatedChild)
	assert.Equal(suite.T(), "child not found", err.Error())
}

func (suite *ChildServiceTestSuite) TestDeleteChildSuccess() {
	// Create a child first
	req := models.CreateChildRequest{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
	}

	createdChild, err := CreateChild(req, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Delete the child
	err = DeleteChild(createdChild.ID)
	assert.NoError(suite.T(), err)

	// Verify child is deleted
	child, err := GetChildByID(createdChild.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), child)
}

func (suite *ChildServiceTestSuite) TestDeleteChildNotFound() {
	err := DeleteChild(999)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "child not found", err.Error())
}

func (suite *ChildServiceTestSuite) TestCheckChildPermissionOwner() {
	// Create a child
	childReq := models.CreateChildRequest{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
	}

	child, err := CreateChild(childReq, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Owner should have all permissions
	hasViewPermission, err := CheckChildPermission(suite.testUser.ID, child.ID, "VIEW")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasViewPermission)

	hasEditPermission, err := CheckChildPermission(suite.testUser.ID, child.ID, "EDIT")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasEditPermission)
}

func (suite *ChildServiceTestSuite) TestCheckChildPermissionNonOwner() {
	// Create another user
	otherUserReq := models.CreateUserRequest{
		Email:     "other@example.com",
		Password:  "password123",
		FirstName: "Other",
		LastName:  "User",
	}

	otherUser, err := CreateUser(otherUserReq)
	assert.NoError(suite.T(), err)

	// Create a child owned by the test user
	childReq := models.CreateChildRequest{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "3rd",
	}

	child, err := CreateChild(childReq, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Other user should not have permissions
	hasViewPermission, err := CheckChildPermission(otherUser.ID, child.ID, "VIEW")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasViewPermission)

	hasEditPermission, err := CheckChildPermission(otherUser.ID, child.ID, "EDIT")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasEditPermission)
}

func (suite *ChildServiceTestSuite) TestCheckChildPermissionNonExistentChild() {
	hasPermission, err := CheckChildPermission(suite.testUser.ID, 999, "VIEW")
	assert.Error(suite.T(), err)
	assert.False(suite.T(), hasPermission)
}

func TestChildServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ChildServiceTestSuite))
}