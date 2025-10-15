package services

import (
	"errors"
	"fmt"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"gorm.io/gorm"
)

// CreateChild creates a new child
func CreateChild(req models.CreateChildRequest, ownerID uint) (*models.Child, error) {
	child := models.Child{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Grade:     req.Grade,
		OwnerID:   ownerID,
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

	child.FirstName = req.FirstName
	child.LastName = req.LastName
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

// GetChildrenWithBookCounts gets children with their book counts for a specific month
func GetChildrenWithBookCounts(userID uint, year int, month int) ([]models.ChildWithBookCountResponse, error) {
	children, err := GetChildrenWithPermission(userID)
	if err != nil {
		return nil, err
	}

	var childrenWithCounts []models.ChildWithBookCountResponse
	
	// Create start and end dates for the month
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	endYear := year
	endMonth := month + 1
	if endMonth > 12 {
		endMonth = 1
		endYear++
	}
	endDate := fmt.Sprintf("%d-%02d-01", endYear, endMonth)

	for _, child := range children {
		var count int64
		config.DB.Model(&models.Book{}).Where("child_id = ? AND date_read >= ? AND date_read < ?", 
			child.ID, startDate, endDate).Count(&count)
		
		childWithCount := models.ChildWithBookCountResponse{
			ID:        child.ID,
			FirstName: child.FirstName,
			LastName:  child.LastName,
			Grade:     child.Grade,
			OwnerID:   child.OwnerID,
			CreatedAt: child.CreatedAt,
			BookCount: int(count),
		}
		childrenWithCounts = append(childrenWithCounts, childWithCount)
	}

	return childrenWithCounts, nil
}

// GetBookCountsForUserChildren gets just book counts for user's children (for month switching)
func GetBookCountsForUserChildren(userID uint, year int, month int) ([]models.BookCountResponse, error) {
	children, err := GetChildrenWithPermission(userID)
	if err != nil {
		return nil, err
	}

	var bookCounts []models.BookCountResponse
	
	// Create start and end dates for the month
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	endYear := year
	endMonth := month + 1
	if endMonth > 12 {
		endMonth = 1
		endYear++
	}
	endDate := fmt.Sprintf("%d-%02d-01", endYear, endMonth)

	for _, child := range children {
		var count int64
		config.DB.Model(&models.Book{}).Where("child_id = ? AND date_read >= ? AND date_read < ?", 
			child.ID, startDate, endDate).Count(&count)
		
		bookCount := models.BookCountResponse{
			ChildID:   child.ID,
			BookCount: int(count),
		}
		bookCounts = append(bookCounts, bookCount)
	}

	return bookCounts, nil
}