package services

import (
	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
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