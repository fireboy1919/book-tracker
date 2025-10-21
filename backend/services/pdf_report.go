package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/booktracker/backend/config"
	"github.com/booktracker/backend/models"
	"github.com/jung-kurt/gofpdf"
)

// BookForPDF represents a book with all the data needed for PDF generation
type BookForPDF struct {
	Title         string
	Author        string
	ISBN          string
	LexileLevel   string
	DateRead      time.Time
	CoverURL      string
	IsPartial     bool
	PartialComment string
	CoverImagePath string // Local path to downloaded cover
	ReadByParent   bool    // Whether book was read by parent vs child
}

// GenerateMonthlyBooksPDF creates a PDF report for a child's books in a specific month
func GenerateMonthlyBooksPDF(childID uint, year int, month int) (string, error) {
	// Get child information
	child, err := GetChildByID(childID)
	if err != nil {
		return "", err
	}

	// Get books for the month
	books, err := getBooksForMonth(childID, year, month)
	if err != nil {
		return "", err
	}

	// Download cover images
	err = downloadCoverImages(books)
	if err != nil {
		return "", err
	}

	// Generate PDF
	pdfPath, err := createPDF(child, books, year, month)
	if err != nil {
		return "", err
	}

	// Clean up cover images
	cleanupCoverImages(books)

	return pdfPath, nil
}

// getBooksForMonth retrieves books for a specific month
func getBooksForMonth(childID uint, year int, month int) ([]*BookForPDF, error) {
	db := config.GetDB()
	
	// Create date range for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
	
	var dbBooks []models.Book
	err := db.Where("child_id = ? AND date_read BETWEEN ? AND ?", childID, startDate, endDate).
		Preload("SharedBook").
		Order("date_read ASC").
		Find(&dbBooks).Error
	if err != nil {
		return nil, err
	}

	// Convert to PDF format
	var books []*BookForPDF
	for _, book := range dbBooks {
		// Parse DateRead string to time.Time
		dateRead, err := time.Parse("2006-01-02", book.DateRead)
		if err != nil {
			// Try alternative format if the first one fails
			dateRead, err = time.Parse("2006-01-02T15:04:05Z07:00", book.DateRead)
			if err != nil {
				// Default to current time if parsing fails
				dateRead = time.Now()
			}
		}
		
		pdfBook := &BookForPDF{
			DateRead:       dateRead,
			IsPartial:      book.IsPartial,
			PartialComment: book.PartialComment,
			LexileLevel:    book.LexileLevel,
			ReadByParent:   book.ReadByParent,
		}

		// Get book details from SharedBook or custom fields
		if book.SharedBook != nil {
			pdfBook.Title = book.SharedBook.Title
			pdfBook.Author = book.SharedBook.Author
			pdfBook.ISBN = book.SharedBook.ISBN
			pdfBook.CoverURL = book.SharedBook.CoverURL
		} else {
			pdfBook.Title = book.CustomTitle
			pdfBook.Author = book.CustomAuthor
			pdfBook.ISBN = book.CustomISBN
			// Custom books don't have cover URLs
		}

		books = append(books, pdfBook)
	}

	return books, nil
}

// downloadCoverImages downloads cover images to temp directory
func downloadCoverImages(books []*BookForPDF) error {
	tempDir := os.TempDir()
	
	for i, book := range books {
		if book.CoverURL == "" {
			continue
		}

		// Create temp file path
		filename := fmt.Sprintf("book_cover_%d.jpg", i)
		filepath := filepath.Join(tempDir, filename)
		
		// Download image
		resp, err := http.Get(book.CoverURL)
		if err != nil {
			continue // Skip if download fails
		}
		defer resp.Body.Close()

		// Create file
		file, err := os.Create(filepath)
		if err != nil {
			continue
		}
		defer file.Close()

		// Copy image data
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			os.Remove(filepath)
			continue
		}

		book.CoverImagePath = filepath
	}
	
	return nil
}

// cleanupCoverImages removes downloaded cover images
func cleanupCoverImages(books []*BookForPDF) {
	for _, book := range books {
		if book.CoverImagePath != "" {
			os.Remove(book.CoverImagePath)
		}
	}
}

// createPDF generates the actual PDF document
func createPDF(child *models.Child, books []*BookForPDF, year int, month int) (string, error) {
	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// Header
	monthName := time.Month(month).String()
	header := fmt.Sprintf("%s %s - %s %d", child.FirstName, child.LastName, monthName, year)
	pdf.Cell(0, 10, header)
	pdf.Ln(10)

	// Calculate reading statistics
	totalBooks := len(books)
	booksReadByChild := 0
	booksReadByParent := 0
	
	for _, book := range books {
		if book.ReadByParent {
			booksReadByParent++
		} else {
			booksReadByChild++
		}
	}

	// Display reading statistics
	pdf.SetFont("Arial", "", 12)
	// Split the text to color the asterisk separately
	beforeStar := fmt.Sprintf("Total books: %d  |  ", totalBooks)
	afterStar := fmt.Sprintf(" Read by %s: %d  |  Read by other: %d", child.FirstName, booksReadByChild, booksReadByParent)
	
	// Print text before star
	pdf.Cell(pdf.GetStringWidth(beforeStar), 8, beforeStar)
	// Print star image
	currentX, currentYPos := pdf.GetXY()
	starSize := 3.0 // 3mm star size
	pdf.ImageOptions("backend/assets/star.png", currentX, currentYPos+1, starSize, starSize,
		false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}, 0, "")
	// Move cursor past the star
	pdf.SetXY(currentX+starSize+0.5, currentYPos)
	// Print remaining text
	pdf.Cell(0, 8, afterStar)
	pdf.Ln(1) // Add line break
	pdf.Ln(14) // Add spacing

	// Page dimensions
	pageWidth, pageHeight := pdf.GetPageSize()
	leftMargin, topMargin, rightMargin, bottomMargin := pdf.GetMargins()
	usableWidth := pageWidth - leftMargin - rightMargin

	// Calculate layout for 4-column display with dynamic row heights
	columnsPerRow := 4
	columnSpacing := 3.0 // Reduced spacing for more column width
	columnWidth := (usableWidth - (columnSpacing * float64(columnsPerRow-1))) / float64(columnsPerRow)
	
	// Start drawing books
	currentY := topMargin + 40 // Start after header and statistics
	
	// Process books in groups of 4 (one row at a time)
	for i := 0; i < len(books); i += columnsPerRow {
		// Get books for this row (up to 4)
		rowBooks := books[i:]
		if len(rowBooks) > columnsPerRow {
			rowBooks = rowBooks[:columnsPerRow]
		}
		
		// Calculate the maximum height needed for this entire row
		maxRowHeight := 30.0 // minimum height
		for _, book := range rowBooks {
			bookHeight := calculateRowHeight(pdf, book, columnWidth)
			if bookHeight > maxRowHeight {
				maxRowHeight = bookHeight
			}
		}
		
		// Check if we need a new page
		if currentY + maxRowHeight > pageHeight - bottomMargin - 30 {
			pdf.AddPage()
			pdf.SetFont("Arial", "B", 16)
			pdf.Cell(0, 10, header)
			pdf.Ln(10)
			
			// Add statistics on new page as well
			pdf.SetFont("Arial", "", 12)
			// Split the text to color the asterisk separately
			beforeStar := fmt.Sprintf("Total books: %d  |  ", totalBooks)
			afterStar := fmt.Sprintf(" Read by %s: %d  |  Read by other: %d", child.FirstName, booksReadByChild, booksReadByParent)
			
			// Print text before star
			pdf.Cell(pdf.GetStringWidth(beforeStar), 8, beforeStar)
			// Print star image
			currentX, currentYPos := pdf.GetXY()
			starSize := 3.0 // 3mm star size
			pdf.ImageOptions("backend/assets/star.png", currentX, currentYPos+1, starSize, starSize,
				false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}, 0, "")
			// Move cursor past the star
			pdf.SetXY(currentX+starSize+0.5, currentYPos)
			// Print remaining text
			pdf.Cell(0, 8, afterStar)
			pdf.Ln(1) // Add line break
			pdf.Ln(14) // Add spacing
			currentY = topMargin + 40
		}
		
		// Draw all books in this row using the same height
		for j, book := range rowBooks {
			columnX := leftMargin + float64(j)*(columnWidth+columnSpacing)
			drawBookColumn(pdf, book, columnX, currentY, columnWidth, maxRowHeight)
		}
		
		// Move to next row
		currentY += maxRowHeight + 2 // Add small spacing between rows
	}

	// Save PDF
	tempDir := os.TempDir()
	pdfPath := filepath.Join(tempDir, fmt.Sprintf("books_report_%s_%s_%s_%d.pdf", 
		child.FirstName, child.LastName, monthName, year))
	
	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return "", err
	}

	return pdfPath, nil
}

// calculateRowHeight determines the height needed for a book entry based on text wrapping
func calculateRowHeight(pdf *gofpdf.Fpdf, book *BookForPDF, columnWidth float64) float64 {
	// Minimum height for cover image area
	minHeight := 30.0
	
	// Calculate text area width (subtract cover area and margins)
	hasImage := book.CoverImagePath != "" && fileExists(book.CoverImagePath)
	var textWidth float64
	if hasImage {
		textWidth = columnWidth - 20.0 - 4.0 // 20mm cover + 4mm margins
	} else {
		textWidth = columnWidth - 4.0 // Just margins
	}
	
	// Calculate height needed for title text (most likely to wrap)
	pdf.SetFont("Arial", "B", 10) // Slightly smaller font for columns
	titleLines := wrapText(pdf, book.Title, textWidth)
	titleHeight := float64(len(titleLines)) * 4.0
	
	// Calculate height needed for author text
	pdf.SetFont("Arial", "", 9)
	authorLines := wrapText(pdf, "by "+book.Author, textWidth)
	authorHeight := float64(len(authorLines)) * 3.5
	
	// Total text height plus spacing
	textHeight := 6.0 + titleHeight + authorHeight + 8.0 // date + title + author + details
	
	// Return the larger of minimum height or calculated text height
	if textHeight > minHeight {
		return textHeight
	}
	return minHeight
}

// drawBookColumn draws a single book in a column with proper text wrapping
func drawBookColumn(pdf *gofpdf.Fpdf, book *BookForPDF, x, y, width, height float64) {
	// Draw column border
	pdf.SetDrawColor(220, 220, 220)
	pdf.Rect(x, y, width, height, "D")
	
	margin := 2.0
	currentY := y + margin
	
	// Check if we have a cover image
	hasImage := book.CoverImagePath != "" && fileExists(book.CoverImagePath)
	
	var textX, textWidth float64
	
	if hasImage {
		// Cover image area (left side, smaller for column layout)
		maxCoverWidth := 20.0
		maxCoverHeight := 25.0
		coverX := x + margin
		coverY := currentY
		
		// Calculate actual image dimensions while maintaining aspect ratio
		actualWidth, actualHeight := calculateImageDimensions(pdf, book.CoverImagePath, maxCoverWidth, maxCoverHeight)
		
		pdf.ImageOptions(book.CoverImagePath, coverX, coverY, actualWidth, actualHeight, 
			false, gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: false}, 0, "")
		
		// Text area (right side) - use maxCoverWidth for consistent text positioning
		textX = coverX + maxCoverWidth + 2
		textWidth = width - maxCoverWidth - margin*2 - 2
	} else {
		// No image - use full width for text
		textX = x + margin
		textWidth = width - margin*2
	}
	
	// Date and status indicators
	pdf.SetXY(textX, currentY)
	pdf.SetFont("Arial", "", 7)
	dateStr := book.DateRead.Format("Jan 2")
	if book.IsPartial {
		dateStr = "PARTIAL - " + dateStr
	}
	
	// Add star for child-read books on the right side
	if book.ReadByParent {
		pdf.CellFormat(textWidth, 3, dateStr, "0", 1, "L", false, 0, "")
	} else {
		// Calculate position for star on far right
		starSize := 2.5 // Smaller star for individual books
		starX := textX + textWidth - starSize - 1 // Position star near right edge
		pdf.CellFormat(textWidth-starSize-1, 3, dateStr, "0", 0, "L", false, 0, "")
		// Draw star image
		pdf.ImageOptions("backend/assets/star.png", starX, currentY-0.5, starSize, starSize,
			false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}, 0, "")
		pdf.Ln(3) // Move to next line
	}
	currentY += 4
	
	// Title with text wrapping
	pdf.SetFont("Arial", "B", 10)
	titleLines := wrapText(pdf, book.Title, textWidth)
	for _, line := range titleLines {
		pdf.SetXY(textX, currentY)
		pdf.CellFormat(textWidth, 4, line, "0", 1, "L", false, 0, "")
		currentY += 4
	}
	
	// Author with text wrapping
	pdf.SetFont("Arial", "", 9)
	authorText := "by " + book.Author
	authorLines := wrapText(pdf, authorText, textWidth)
	for _, line := range authorLines {
		pdf.SetXY(textX, currentY)
		pdf.CellFormat(textWidth, 3.5, line, "0", 1, "L", false, 0, "")
		currentY += 3.5
	}
	
	// Additional details (Lexile level if available)
	if book.LexileLevel != "" {
		pdf.SetXY(textX, currentY)
		pdf.CellFormat(textWidth, 3, "Lexile: "+book.LexileLevel, "0", 1, "L", false, 0, "")
	}
}

// Helper functions
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// calculateImageDimensions calculates scaled image dimensions that fit within maxWidth x maxHeight while maintaining aspect ratio
func calculateImageDimensions(pdf *gofpdf.Fpdf, imagePath string, maxWidth, maxHeight float64) (float64, float64) {
	// Get image info from gofpdf
	info := pdf.RegisterImageOptions(imagePath, gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: false})
	if info == nil {
		// If we can't get image info, return default size
		return maxWidth, maxHeight
	}
	
	// Get original dimensions in points (gofpdf uses points internally)
	originalWidth, originalHeight := info.Extent()
	
	if originalWidth <= 0 || originalHeight <= 0 {
		return maxWidth, maxHeight
	}
	
	// Convert to mm (gofpdf's default unit for this PDF)
	// 1 point = 0.352778 mm
	pointToMM := 0.352778
	origWidthMM := originalWidth * pointToMM
	origHeightMM := originalHeight * pointToMM
	
	// Calculate aspect ratio
	aspectRatio := origWidthMM / origHeightMM
	
	// Scale to fit within max dimensions while maintaining aspect ratio
	var scaledWidth, scaledHeight float64
	
	if aspectRatio > (maxWidth / maxHeight) {
		// Image is wider relative to container - constrain by width
		scaledWidth = maxWidth
		scaledHeight = maxWidth / aspectRatio
	} else {
		// Image is taller relative to container - constrain by height
		scaledHeight = maxHeight
		scaledWidth = maxHeight * aspectRatio
	}
	
	return scaledWidth, scaledHeight
}

// wrapText breaks text into lines that fit within the specified width
func wrapText(pdf *gofpdf.Fpdf, text string, maxWidth float64) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	
	var lines []string
	currentLine := ""
	
	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word
		
		// Check if this line fits
		lineWidth := pdf.GetStringWidth(testLine)
		if lineWidth <= maxWidth {
			currentLine = testLine
		} else {
			// Line too long, save current line and start new one
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
			
			// If even single word is too long, truncate it
			if pdf.GetStringWidth(currentLine) > maxWidth {
				// Try to fit as much as possible
				for len(currentLine) > 0 {
					if pdf.GetStringWidth(currentLine+"...") <= maxWidth {
						currentLine += "..."
						break
					}
					currentLine = currentLine[:len(currentLine)-1]
				}
			}
		}
	}
	
	// Add the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	if len(lines) == 0 {
		return []string{""}
	}
	
	return lines
}