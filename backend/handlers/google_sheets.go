package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/middleware"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
)

// CreatePersonalizedGoogleSheet creates a personalized Google Sheets template for the teacher
func CreatePersonalizedGoogleSheet(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid class ID",
		})
		return
	}

	// Verify user has access to this class (teacher or admin)
	isAdmin, _ := c.Get("isAdmin")
	isTeacher, _ := c.Get("isTeacher")
	
	if !isAdmin.(bool) && !isTeacher.(bool) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Message: "Only teachers and admins can create Google Sheets templates",
		})
		return
	}

	// Check if user has access to this specific class
	if !isAdmin.(bool) {
		var membership models.ClassMembership
		if err := config.GetDB().Where("class_id = ? AND user_id = ? AND role = ?", classID, userID, "TEACHER").First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Message: "You don't have access to this class",
			})
			return
		}
	}

	// Get teacher invitation data
	invitationService := services.NewStudentInvitationService(config.GetDB())
	data, err := invitationService.GenerateTeacherInvitationData(userID, uint(classID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to generate invitation data: " + err.Error(),
		})
		return
	}

	// Create the Google Apps Script URL that will create a personalized copy
	scriptURL := "https://script.google.com/macros/s/YOUR_APPS_SCRIPT_ID/exec"
	
	// Return the URL with embedded teacher data
	c.JSON(http.StatusOK, gin.H{
		"google_sheets_url": scriptURL,
		"teacher_data": data,
		"setup_url": fmt.Sprintf("%s?teacherId=%d&classId=%d&invitationKey=%s&className=%s", 
			scriptURL, 
			data.TeacherID, 
			data.ClassID, 
			url.QueryEscape(data.InvitationKey),
			url.QueryEscape(data.ClassName)),
	})
}