package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/booktracker/backend/middleware"
	"github.com/booktracker/backend/models"
	"github.com/booktracker/backend/services"
	"github.com/gin-gonic/gin"
)

// GenerateMonthlyPDFReport generates a PDF report for a child's books in a specific month
func GenerateMonthlyPDFReport(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Message: "User not found",
		})
		return
	}

	// Parse child ID
	childIDParam := c.Param("childId")
	childID, err := strconv.ParseUint(childIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid child ID",
		})
		return
	}

	// Parse year and month from query params
	yearParam := c.Query("year")
	monthParam := c.Query("month")
	
	if yearParam == "" || monthParam == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Year and month parameters are required",
		})
		return
	}

	year, err := strconv.Atoi(yearParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid year parameter",
		})
		return
	}

	month, err := strconv.Atoi(monthParam)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Invalid month parameter",
		})
		return
	}

	// Check permission to access this child
	hasPermission, err := services.CheckChildPermission(userID, uint(childID), "VIEW")
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

	// Generate PDF
	pdfPath, err := services.GenerateMonthlyBooksPDF(uint(childID), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to generate PDF: " + err.Error(),
		})
		return
	}

	// Ensure cleanup after serving
	defer func() {
		os.Remove(pdfPath)
	}()

	// Get file info
	fileInfo, err := os.Stat(pdfPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Failed to access generated PDF",
		})
		return
	}

	// Set headers for PDF download
	filename := filepath.Base(pdfPath)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// Serve the PDF file
	c.File(pdfPath)
}