package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/middleware"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
)

// GetTeacherInvitationData generates the invitation key and data for Google Sheets
func GetTeacherInvitationData(c *gin.Context) {
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
			Message: "Only teachers and admins can generate invitation data",
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

	invitationService := services.NewStudentInvitationService(config.GetDB())
	data, err := invitationService.GenerateTeacherInvitationData(userID, uint(classID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to generate invitation data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GenerateInvitationToken creates an encrypted token for a specific student in a class
func GenerateInvitationToken(c *gin.Context) {
	// Get class ID from URL path
	classIDStr := c.Param("id")
	classID, err := strconv.ParseUint(classIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid class ID",
		})
		return
	}

	var req struct {
		StudentName string `json:"student_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Create payload with compact format (no class ID needed - comes from URL)
	payload := models.StudentInvitationPayload{
		StudentName: req.StudentName,
		Timestamp:   time.Now().Unix(),
	}

	invitationService := services.NewStudentInvitationService(config.GetDB())
	token, err := invitationService.EncryptInvitationDataForClass(payload, uint(classID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to generate token: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"url":   "https://booktracker.app/invite/" + token,
	})
}

// RedeemStudentInvitation processes a student invitation token
func RedeemStudentInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Missing invitation token",
		})
		return
	}

	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		// Store the token in session/cookie for after login
		c.SetCookie("pending_invitation", token, 3600, "/", "", false, true)
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Please log in to redeem this invitation",
			"token":   token,
		})
		return
	}

	invitationService := services.NewStudentInvitationService(config.GetDB())
	child, err := invitationService.RedeemInvitation(token, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Failed to redeem invitation: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student invitation redeemed successfully",
		"child":   child,
	})
}

// GetStudentInvitationDetails shows invitation information before redemption
func GetStudentInvitationDetails(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Missing invitation token",
		})
		return
	}

	invitationService := services.NewStudentInvitationService(config.GetDB())
	payload, err := invitationService.DecryptInvitationToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid or expired invitation: " + err.Error(),
		})
		return
	}

	// Get class information
	var class models.Class
	if err := config.GetDB().First(&class, payload.ClassID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: "Class not found",
		})
		return
	}

	// Get teacher information from class membership
	var membership models.ClassMembership
	var teacher models.User
	if err := config.GetDB().Where("class_id = ? AND role = ?", payload.ClassID, "TEACHER").First(&membership).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: "Teacher not found for this class",
		})
		return
	}
	
	if err := config.GetDB().First(&teacher, membership.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Message: "Teacher not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"student_name": payload.StudentName,
		"teacher_name": teacher.FirstName + " " + teacher.LastName,
		"class_name":   class.Name,
		"valid":        true,
	})
}

// CheckPendingInvitation checks if user has a pending invitation cookie
func CheckPendingInvitation(c *gin.Context) {
	_, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not authenticated",
		})
		return
	}

	token, err := c.Cookie("pending_invitation")
	if err != nil || token == "" {
		c.JSON(http.StatusOK, gin.H{
			"has_pending": false,
		})
		return
	}

	// Try to get invitation details
	invitationService := services.NewStudentInvitationService(config.GetDB())
	payload, err := invitationService.DecryptInvitationToken(token)
	if err != nil {
		// Clear invalid cookie
		c.SetCookie("pending_invitation", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{
			"has_pending": false,
		})
		return
	}

	// Get class info
	var class models.Class
	config.GetDB().First(&class, payload.ClassID)
	
	// Get teacher info from class membership
	var membership models.ClassMembership
	var teacher models.User
	config.GetDB().Where("class_id = ? AND role = ?", payload.ClassID, "TEACHER").First(&membership)
	config.GetDB().First(&teacher, membership.UserID)

	c.JSON(http.StatusOK, gin.H{
		"has_pending":    true,
		"token":          token,
		"student_name":   payload.StudentName,
		"teacher_name":   teacher.FirstName + " " + teacher.LastName,
		"class_name":     class.Name,
	})
}

// RedeemPendingInvitation redeems a pending invitation from cookie
func RedeemPendingInvitation(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not authenticated",
		})
		return
	}

	token, err := c.Cookie("pending_invitation")
	if err != nil || token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "No pending invitation found",
		})
		return
	}

	invitationService := services.NewStudentInvitationService(config.GetDB())
	child, err := invitationService.RedeemInvitation(token, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Failed to redeem invitation: " + err.Error(),
		})
		return
	}

	// Clear the cookie
	c.SetCookie("pending_invitation", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Student invitation redeemed successfully",
		"child":   child,
	})
}

