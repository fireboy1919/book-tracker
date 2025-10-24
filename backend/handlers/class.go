package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
)

// CreateClass creates a new class
func CreateClass(c *gin.Context) {
	var req models.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	class, err := classService.CreateClass(userID.(uint), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, class)
}

// GetClasses returns all classes for admin or classes where user is a member
func GetClasses(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	classes, err := classService.GetClasses(userID.(uint), isAdmin.(bool))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}

// GetClass returns a specific class with its members and children
func GetClass(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid class ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	class, err := classService.GetClass(uint(classID), userID.(uint), isAdmin.(bool))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Class not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, class)
}

// UpdateClass updates a class
func UpdateClass(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid class ID"})
		return
	}

	var req models.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	class, err := classService.UpdateClass(uint(classID), userID.(uint), isAdmin.(bool), req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Class not found"})
			return
		}
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, class)
}

// DeleteClass deletes a class
func DeleteClass(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid class ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	err = classService.DeleteClass(uint(classID), userID.(uint), isAdmin.(bool))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Class not found"})
			return
		}
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Class deleted successfully"})
}

// AddClassMember adds a user to a class with a specific role
func AddClassMember(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid class ID"})
		return
	}

	var req models.AddClassMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	membership, err := classService.AddClassMember(uint(classID), userID.(uint), isAdmin.(bool), req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Class or user not found"})
			return
		}
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, membership)
}

// RemoveClassMember removes a user from a class
func RemoveClassMember(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid class ID"})
		return
	}

	memberIDStr := c.Param("userId")
	memberID, err := strconv.ParseUint(memberIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid user ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	err = classService.RemoveClassMember(uint(classID), uint(memberID), userID.(uint), isAdmin.(bool))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Class membership not found"})
			return
		}
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed from class successfully"})
}

// GetClassStudents returns all students in a class, sorted alphabetically
func GetClassStudents(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid class ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	students, err := classService.GetClassStudents(uint(classID), userID.(uint), isAdmin.(bool))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Class not found"})
			return
		}
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}

// RemoveChildFromClass removes a child from a class
func RemoveChildFromClass(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid class ID"})
		return
	}

	childIDStr := c.Param("childId")
	childID, err := strconv.ParseUint(childIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid child ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	err = classService.RemoveChildFromClass(uint(childID), uint(classID), userID.(uint), isAdmin.(bool))
	if err != nil {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Child removed from class successfully"})
}

// GetClassTeachers returns all teachers in a class, sorted alphabetically
func GetClassTeachers(c *gin.Context) {
	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid class ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	teachers, err := classService.GetClassTeachers(uint(classID), userID.(uint), isAdmin.(bool))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Class not found"})
			return
		}
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, teachers)
}

// AssignChildToClass assigns a child to a class
func AssignChildToClass(c *gin.Context) {
	var req models.AssignChildToClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User role not found"})
		return
	}

	classService := services.NewClassService(config.GetDB())
	child, err := classService.AssignChildToClass(req.ChildID, req.ClassID, userID.(uint), isAdmin.(bool))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Child or class not found"})
			return
		}
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, child)
}

// GetAvailableClasses returns all classes available for assignment (for parents)
func GetAvailableClasses(c *gin.Context) {
	classService := services.NewClassService(config.GetDB())
	classes, err := classService.GetAvailableClasses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}

// SearchStudents searches for students by name for teachers to add to classes
func SearchStudents(c *gin.Context) {
	fmt.Printf("SearchStudents handler called\n")
	userID, exists := c.Get("userID")
	if !exists {
		fmt.Printf("User not authenticated in SearchStudents\n")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User not authenticated"})
		return
	}

	isAdminVal, _ := c.Get("isAdmin")
	query := c.Query("q")
	fmt.Printf("SearchStudents query: %s, userID: %v, isAdmin: %v\n", query, userID, isAdminVal)
	if query == "" {
		fmt.Printf("Missing query parameter in SearchStudents\n")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Query parameter 'q' is required"})
		return
	}

	// Safely convert isAdmin to bool
	isAdmin := false
	if isAdminVal != nil {
		if admin, ok := isAdminVal.(bool); ok {
			isAdmin = admin
		}
	}

	classService := services.NewClassService(config.GetDB())
	students, err := classService.SearchStudents(query, userID.(uint), isAdmin)
	if err != nil {
		fmt.Printf("SearchStudents service error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	fmt.Printf("SearchStudents returning %d students\n", len(students))
	c.JSON(http.StatusOK, students)
}