package services

import (
	"github.com/booktracker/api/config"
	"github.com/booktracker/api/models"
)

// CreatePermission creates a new permission for a user to access a child's data
func CreatePermission(userID, childID uint, permissionType string) error {
	// Check if permission already exists
	var existingPermission models.Permission
	result := config.DB.Where("user_id = ? AND child_id = ?", userID, childID).First(&existingPermission)
	
	if result.Error == nil {
		// Permission exists, update it
		existingPermission.PermissionType = permissionType
		return config.DB.Save(&existingPermission).Error
	}

	// Create new permission
	permission := models.Permission{
		UserID:         userID,
		ChildID:        childID,
		PermissionType: permissionType,
	}

	return config.DB.Create(&permission).Error
}

// GetPermissionsByUser gets all permissions for a specific user
func GetPermissionsByUser(userID uint) ([]models.Permission, error) {
	var permissions []models.Permission
	result := config.DB.Where("user_id = ?", userID).Find(&permissions)
	return permissions, result.Error
}

// GetPermissionsByChild gets all permissions for a specific child
func GetPermissionsByChild(childID uint) ([]models.Permission, error) {
	var permissions []models.Permission
	result := config.DB.Where("child_id = ?", childID).Find(&permissions)
	return permissions, result.Error
}


// DeletePermission removes a permission
func DeletePermission(userID, childID uint) error {
	return config.DB.Where("user_id = ? AND child_id = ?", userID, childID).Delete(&models.Permission{}).Error
}

// GetPermissionByID gets a permission by ID
func GetPermissionByID(permissionID uint) (*models.Permission, error) {
	var permission models.Permission
	result := config.DB.First(&permission, permissionID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &permission, nil
}

// DeletePermissionByID removes a permission by ID
func DeletePermissionByID(permissionID uint) error {
	return config.DB.Delete(&models.Permission{}, permissionID).Error
}

// CreateOrUpdatePermission is an alias for CreatePermission which already handles updates
func CreateOrUpdatePermission(userID, childID uint, permissionType string) error {
	return CreatePermission(userID, childID, permissionType)
}