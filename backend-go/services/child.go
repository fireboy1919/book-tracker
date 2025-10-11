package services

import (
	"errors"

	"github.com/booktracker/backend-go/config"
	"github.com/booktracker/backend-go/models"
	"gorm.io/gorm"
)

// CreateChild creates a new child
func CreateChild(req models.CreateChildRequest, ownerID uint) (*models.Child, error) {
	child := models.Child{
		Name:    req.Name,
		Grade:   req.Grade,
		OwnerID: ownerID,
	}

	result := config.DB.Create(&child)
	if result.Error != nil {
		return nil, result.Error
	}

	return &child, nil
}

// GetChildByID gets a child by ID
func GetChildByID(id uint) (*models.Child, error) {
	var child models.Child
	result := config.DB.Preload("Owner").First(&child, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("child not found")
		}
		return nil, result.Error
	}
	return &child, nil
}

// GetChildrenByOwner gets all children owned by a user
func GetChildrenByOwner(ownerID uint) ([]models.Child, error) {
	var children []models.Child
	result := config.DB.Where("owner_id = ?", ownerID).Find(&children)
	if result.Error != nil {
		return nil, result.Error
	}
	return children, nil
}

// GetChildrenWithPermission gets children that a user has permission to view
func GetChildrenWithPermission(userID uint) ([]models.Child, error) {
	var children []models.Child
	
	// Get children owned by user or children user has permissions for
	result := config.DB.Raw(`
		SELECT DISTINCT c.* FROM children c 
		LEFT JOIN permissions p ON c.id = p.child_id 
		WHERE c.owner_id = ? OR p.user_id = ?
	`, userID, userID).Scan(&children)
	
	if result.Error != nil {
		return nil, result.Error
	}
	return children, nil
}

// UpdateChild updates a child
func UpdateChild(id uint, req models.UpdateChildRequest) (*models.Child, error) {
	var child models.Child
	result := config.DB.First(&child, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("child not found")
		}
		return nil, result.Error
	}

	child.Name = req.Name
	child.Grade = req.Grade

	result = config.DB.Save(&child)
	if result.Error != nil {
		return nil, result.Error
	}

	return &child, nil
}

// DeleteChild deletes a child
func DeleteChild(id uint) error {
	result := config.DB.Delete(&models.Child{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("child not found")
	}
	return nil
}

// CheckChildPermission checks if a user has permission to access a child
func CheckChildPermission(userID, childID uint, permissionType string) (bool, error) {
	var child models.Child
	result := config.DB.First(&child, childID)
	if result.Error != nil {
		return false, result.Error
	}

	// Owner has all permissions
	if child.OwnerID == userID {
		return true, nil
	}

	// Check explicit permissions
	var permission models.Permission
	result = config.DB.Where("user_id = ? AND child_id = ? AND permission_type = ?", userID, childID, permissionType).First(&permission)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Also check for EDIT permission which includes VIEW
			if permissionType == "VIEW" {
				result = config.DB.Where("user_id = ? AND child_id = ? AND permission_type = ?", userID, childID, "EDIT").First(&permission)
				if result.Error != nil {
					return false, nil
				}
				return true, nil
			}
			return false, nil
		}
		return false, result.Error
	}

	return true, nil
}