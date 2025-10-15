package handlers

import (
	"net/http"
	"strconv"

	"github.com/booktracker/backend/middleware"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
	"github.com/gin-gonic/gin"
)

// CreateChild handles creating a new child
func CreateChild(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	var req models.CreateChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	child, err := services.CreateChild(req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to create child: " + err.Error(),
		})
		return
	}

	childResponse := models.ChildResponse{
		ID:        child.ID,
		FirstName: child.FirstName,
		LastName:  child.LastName,
		Grade:     child.Grade,
		OwnerID:   child.OwnerID,
		CreatedAt: child.CreatedAt,
	}

	c.JSON(http.StatusCreated, childResponse)
}

// GetChildren handles getting children for current user
func GetChildren(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	children, err := services.GetChildrenWithPermission(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to get children: " + err.Error(),
		})
		return
	}

	var childResponses []models.ChildResponse
	for _, child := range children {
		childResponses = append(childResponses, models.ChildResponse{
			ID:        child.ID,
			FirstName: child.FirstName,
			LastName:  child.LastName,
			Grade:     child.Grade,
			OwnerID:   child.OwnerID,
			CreatedAt: child.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, childResponses)
}

// GetChildByID handles getting a child by ID
func GetChildByID(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid child ID",
		})
		return
	}

	// Check permission
	hasPermission, err := services.CheckChildPermission(userID, uint(id), "VIEW")
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

	child, err := services.GetChildByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	childResponse := models.ChildResponse{
		ID:        child.ID,
		FirstName: child.FirstName,
		LastName:  child.LastName,
		Grade:     child.Grade,
		OwnerID:   child.OwnerID,
		CreatedAt: child.CreatedAt,
	}

	c.JSON(http.StatusOK, childResponse)
}

// UpdateChild handles updating a child
func UpdateChild(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid child ID",
		})
		return
	}

	// Check permission
	hasPermission, err := services.CheckChildPermission(userID, uint(id), "EDIT")
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

	var req models.UpdateChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	child, err := services.UpdateChild(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	childResponse := models.ChildResponse{
		ID:        child.ID,
		FirstName: child.FirstName,
		LastName:  child.LastName,
		Grade:     child.Grade,
		OwnerID:   child.OwnerID,
		CreatedAt: child.CreatedAt,
	}

	c.JSON(http.StatusOK, childResponse)
}

// DeleteChild handles deleting a child
func DeleteChild(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid child ID",
		})
		return
	}

	// Check permission (only owner can delete)
	child, err := services.GetChildByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	if child.OwnerID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Only the owner can delete a child",
		})
		return
	}

	err = services.DeleteChild(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// InviteUser handles inviting a user to access a child's data
func InviteUser(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	childIDParam := c.Param("id")
	childID, err := strconv.ParseUint(childIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid child ID",
		})
		return
	}

	var req models.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if current user owns the child
	child, err := services.GetChildByID(uint(childID))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: "Child not found",
		})
		return
	}

	if child.OwnerID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Only the owner can invite users to access this child",
		})
		return
	}

	// Check if user exists
	targetUser, err := services.GetUserByEmail(req.Email)
	if err != nil {
		// User doesn't exist - create pending invitation
		invitation, err := services.CreatePendingInvitation(req.Email, uint(childID), req.PermissionType, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to create invitation: " + err.Error(),
			})
			return
		}

		// Send invitation email
		currentUser, err := services.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to get current user: " + err.Error(),
			})
			return
		}

		err = services.SendInvitationEmail(req.Email, invitation.Token, currentUser, child)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Message: "Failed to send invitation email: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Invitation sent to " + req.Email + ". They will need to register to access the child's data."})
		return
	}

	// User exists - check if they already have permission
	hasExistingPermission, err := services.CheckChildPermission(targetUser.ID, uint(childID), "VIEW")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to check existing permissions: " + err.Error(),
		})
		return
	}

	if hasExistingPermission {
		// Update existing permission if it's different
		// (Implementation would need a service to update permissions)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "User already has access to this child",
		})
		return
	}

	// Create permission for the existing user
	err = services.CreatePermission(targetUser.ID, uint(childID), req.PermissionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to create permission: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission granted to existing user"})
}

// GetChildrenWithBookCounts handles getting children with their book counts for a specific month
func GetChildrenWithBookCounts(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	// Required query parameters
	year := c.Query("year")
	month := c.Query("month")
	
	if year == "" || month == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Year and month parameters are required",
		})
		return
	}

	yearInt, yearErr := strconv.Atoi(year)
	monthInt, monthErr := strconv.Atoi(month)
	
	if yearErr != nil || monthErr != nil || monthInt < 1 || monthInt > 12 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid year or month parameter",
		})
		return
	}

	childrenWithCounts, err := services.GetChildrenWithBookCounts(userID, yearInt, monthInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to get children with book counts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, childrenWithCounts)
}

// GetBookCountsForChildren handles getting just book counts for existing children (month switching)
func GetBookCountsForChildren(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	// Required query parameters
	year := c.Query("year")
	month := c.Query("month")
	
	if year == "" || month == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Year and month parameters are required",
		})
		return
	}

	yearInt, yearErr := strconv.Atoi(year)
	monthInt, monthErr := strconv.Atoi(month)
	
	if yearErr != nil || monthErr != nil || monthInt < 1 || monthInt > 12 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid year or month parameter",
		})
		return
	}

	bookCounts, err := services.GetBookCountsForUserChildren(userID, yearInt, monthInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to get book counts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, bookCounts)
}