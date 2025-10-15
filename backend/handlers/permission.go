package handlers

import (
	"net/http"
	"strconv"

	"github.com/booktracker/backend/middleware"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
	"github.com/gin-gonic/gin"
)

// GetPermissionsByChild handles getting all permissions for a specific child
func GetPermissionsByChild(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	childIDParam := c.Param("childId")
	childID, err := strconv.ParseUint(childIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid child ID",
		})
		return
	}

	// Check if user has EDIT permission for this child (only owners and editors can see permissions)
	hasPermission, err := services.CheckChildPermission(userID, uint(childID), "EDIT")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to check permission: " + err.Error(),
		})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	permissions, err := services.GetPermissionsByChild(uint(childID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to get permissions: " + err.Error(),
		})
		return
	}

	var permissionResponses []models.PermissionResponse
	for _, permission := range permissions {
		// Get user details for each permission
		user, err := services.GetUserByID(permission.UserID)
		if err == nil {
			permissionResponses = append(permissionResponses, models.PermissionResponse{
				ID:             permission.ID,
				UserID:         permission.UserID,
				ChildID:        permission.ChildID,
				PermissionType: permission.PermissionType,
				CreatedAt:      permission.CreatedAt,
				User: &models.UserResponse{
					ID:        user.ID,
					Email:     user.Email,
					FirstName: user.FirstName,
					LastName:  user.LastName,
				},
			})
		}
	}

	c.JSON(http.StatusOK, permissionResponses)
}

// DeletePermissionByID handles deleting a specific permission
func DeletePermissionByID(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	permissionIDParam := c.Param("id")
	permissionID, err := strconv.ParseUint(permissionIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid permission ID",
		})
		return
	}

	// Get the permission to check child ownership
	permission, err := services.GetPermissionByID(uint(permissionID))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: "Permission not found",
		})
		return
	}

	// Check if user has EDIT permission for this child
	hasPermission, err := services.CheckChildPermission(userID, permission.ChildID, "EDIT")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to check permission: " + err.Error(),
		})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Access denied",
		})
		return
	}

	err = services.DeletePermissionByID(uint(permissionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to delete permission: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}