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
		Name: "Test Child",
		Age:  8,
	}

	child, err := CreateChild(req, suite.testUser.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), child)
	assert.Equal(suite.T(), "Test Child", child.Name)
	assert.Equal(suite.T(), 8, child.Age)
	assert.Equal(suite.T(), suite.testUser.ID, child.OwnerID)
}

func (suite *ChildServiceTestSuite) TestGetChildByIDSuccess() {
	// Create a child first
	req := models.CreateChildRequest{
		Name: "Test Child",
		Age:  8,
	}

	createdChild, err := CreateChild(req, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Get child by ID
	child, err := GetChildByID(createdChild.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), child)
	assert.Equal(suite.T(), createdChild.ID, child.ID)
	assert.Equal(suite.T(), "Test Child", child.Name)
	assert.Equal(suite.T(), 8, child.Age)
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
		Name: "Child One",
		Age:  7,
	}

	child2Req := models.CreateChildRequest{
		Name: "Child Two",
		Age:  9,
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
	names := make([]string, len(children))
	for i, child := range children {
		names[i] = child.Name
	}
	assert.Contains(suite.T(), names, "Child One")
	assert.Contains(suite.T(), names, "Child Two")
}

func (suite *ChildServiceTestSuite) TestGetChildrenWithPermission() {
	// Create a child
	childReq := models.CreateChildRequest{
		Name: "Test Child",
		Age:  8,
	}

	_, err := CreateChild(childReq, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Get children with permission (owner should have access)
	children, err := GetChildrenWithPermission(suite.testUser.ID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), children, 1)
	assert.Equal(suite.T(), "Test Child", children[0].Name)
}

func (suite *ChildServiceTestSuite) TestUpdateChildSuccess() {
	// Create a child first
	createReq := models.CreateChildRequest{
		Name: "Original Name",
		Age:  7,
	}

	createdChild, err := CreateChild(createReq, suite.testUser.ID)
	assert.NoError(suite.T(), err)

	// Update the child
	updateReq := models.UpdateChildRequest{
		Name: "Updated Name",
		Age:  8,
	}

	updatedChild, err := UpdateChild(createdChild.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedChild)
	assert.Equal(suite.T(), "Updated Name", updatedChild.Name)
	assert.Equal(suite.T(), 8, updatedChild.Age)
	assert.Equal(suite.T(), createdChild.ID, updatedChild.ID)
}

func (suite *ChildServiceTestSuite) TestUpdateChildNotFound() {
	updateReq := models.UpdateChildRequest{
		Name: "Updated Name",
		Age:  8,
	}

	updatedChild, err := UpdateChild(999, updateReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), updatedChild)
	assert.Equal(suite.T(), "child not found", err.Error())
}

func (suite *ChildServiceTestSuite) TestDeleteChildSuccess() {
	// Create a child first
	req := models.CreateChildRequest{
		Name: "Test Child",
		Age:  8,
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
		Name: "Test Child",
		Age:  8,
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
		Name: "Test Child",
		Age:  8,
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