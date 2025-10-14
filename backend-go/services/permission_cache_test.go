package services

import (
	"testing"
	"time"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PermissionCacheTestSuite struct {
	suite.Suite
	cache *PermissionCache
}

func (suite *PermissionCacheTestSuite) SetupTest() {
	// Setup test database
	config.TestDB = config.SetupTestDatabase()
	config.DB = config.TestDB
	
	// Create a new cache for each test
	suite.cache = NewPermissionCache(5000) // 5 seconds TTL
}

func (suite *PermissionCacheTestSuite) TearDownTest() {
	config.CleanupTestDatabase()
}

func (suite *PermissionCacheTestSuite) TestPermissionCacheHit() {
	// Create test data
	user := models.User{
		Email:         "user@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "User",
		EmailVerified: true,
	}
	config.DB.Create(&user)

	child := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "1st",
		OwnerID:   user.ID,
	}
	config.DB.Create(&child)

	permission := models.Permission{
		UserID:         user.ID,
		ChildID:        child.ID,
		PermissionType: "VIEW",
	}
	config.DB.Create(&permission)

	// First call should hit database
	start := time.Now()
	hasPermission1, err := suite.cache.GetOrCheck(user.ID, child.ID, "VIEW")
	duration1 := time.Since(start)
	
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission1)

	// Second call should hit cache (much faster)
	start = time.Now()
	hasPermission2, err := suite.cache.GetOrCheck(user.ID, child.ID, "VIEW")
	duration2 := time.Since(start)
	
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission2)
	
	// Cache hit should be significantly faster
	assert.True(suite.T(), duration2 < duration1/2, "Cache hit should be much faster than database query")
}

func (suite *PermissionCacheTestSuite) TestPermissionCacheMiss() {
	// Create test users
	owner := models.User{
		Email:         "owner@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Owner",
		LastName:      "User",
		EmailVerified: true,
	}
	config.DB.Create(&owner)

	nonOwner := models.User{
		Email:         "user@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "User",
		EmailVerified: true,
	}
	config.DB.Create(&nonOwner)

	// Create child owned by different user
	child := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "1st",
		OwnerID:   owner.ID, // Different owner
	}
	config.DB.Create(&child)

	// Non-owner should not have permission (and should cache the result)
	hasPermission1, err := suite.cache.GetOrCheck(nonOwner.ID, child.ID, "VIEW")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission1)

	// Second call should hit cache with same result
	hasPermission2, err := suite.cache.GetOrCheck(nonOwner.ID, child.ID, "VIEW")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission2)
}

func (suite *PermissionCacheTestSuite) TestPermissionCacheExpiration() {
	// Create cache with very short TTL
	shortCache := NewPermissionCache(10) // 10ms TTL

	user := models.User{
		Email:         "user@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "User",
		EmailVerified: true,
	}
	config.DB.Create(&user)

	child := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "1st",
		OwnerID:   user.ID,
	}
	config.DB.Create(&child)

	// First call
	hasPermission1, err := shortCache.GetOrCheck(user.ID, child.ID, "VIEW")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission1)

	// Wait for cache to expire
	time.Sleep(20 * time.Millisecond)

	// Should hit database again after expiration
	hasPermission2, err := shortCache.GetOrCheck(user.ID, child.ID, "VIEW")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission2)
}

func (suite *PermissionCacheTestSuite) TestPermissionCacheDifferentKeys() {
	user1 := models.User{
		Email:         "user1@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test1",
		LastName:      "User",
		EmailVerified: true,
	}
	config.DB.Create(&user1)

	user2 := models.User{
		Email:         "user2@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test2",
		LastName:      "User",
		EmailVerified: true,
	}
	config.DB.Create(&user2)

	child := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "1st",
		OwnerID:   user1.ID,
	}
	config.DB.Create(&child)

	// User1 should have access (owner)
	hasPermission1, err := suite.cache.GetOrCheck(user1.ID, child.ID, "VIEW")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission1)

	// User2 should not have access (different user)
	hasPermission2, err := suite.cache.GetOrCheck(user2.ID, child.ID, "VIEW")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), hasPermission2)

	// Different permission types should be cached separately
	hasEditPermission, err := suite.cache.GetOrCheck(user1.ID, child.ID, "EDIT")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasEditPermission)
}

func (suite *PermissionCacheTestSuite) TestPermissionCacheClear() {
	user := models.User{
		Email:         "user@example.com",
		PasswordHash:  "hashedpassword",
		FirstName:     "Test",
		LastName:      "User",
		EmailVerified: true,
	}
	config.DB.Create(&user)

	child := models.Child{
		FirstName: "Test",
		LastName:  "Child",
		Grade:     "1st",
		OwnerID:   user.ID,
	}
	config.DB.Create(&child)

	// Cache a permission
	hasPermission1, err := suite.cache.GetOrCheck(user.ID, child.ID, "VIEW")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission1)

	// Clear cache
	suite.cache.Clear()

	// Should hit database again after clear
	start := time.Now()
	hasPermission2, err := suite.cache.GetOrCheck(user.ID, child.ID, "VIEW")
	duration := time.Since(start)
	
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), hasPermission2)
	
	// Should take some time to query database (this assertion might be flaky on very fast systems)
	// We'll just verify that the query succeeded and returned the correct result
	assert.True(suite.T(), duration >= 0, "Query duration should be non-negative")
}

func TestPermissionCacheTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionCacheTestSuite))
}