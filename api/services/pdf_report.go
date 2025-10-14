package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/booktracker/api/config"
	"github.com/booktracker/api/models"
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
	pdf.Ln(15)

	// Page dimensions
	pageWidth, pageHeight := pdf.GetPageSize()
	leftMargin, topMargin, rightMargin, bottomMargin := pdf.GetMargins()
	usableWidth := pageWidth - leftMargin - rightMargin
	usableHeight := pageHeight - topMargin - bottomMargin - 25 // Reserve space for header

	// Calculate layout: 4 columns, 8 rows = 32 books per page
	cols := 4
	rows := 8
	cellWidth := usableWidth / float64(cols)
	cellHeight := usableHeight / float64(rows)

	// Draw books in grid
	for i, book := range books {
		if i > 0 && i%32 == 0 {
			// Add new page every 32 books
			pdf.AddPage()
			pdf.SetFont("Arial", "B", 16)
			pdf.Cell(0, 10, header)
			pdf.Ln(15)
		}

		// Calculate position
		bookIndex := i % 32
		col := bookIndex % cols
		row := bookIndex / cols

		x := leftMargin + float64(col)*cellWidth
		y := topMargin + 25 + float64(row)*cellHeight // 25 for header space

		drawBookCell(pdf, book, x, y, cellWidth, cellHeight)
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

// drawBookCell draws a single book in the PDF grid
func drawBookCell(pdf *gofpdf.Fpdf, book *BookForPDF, x, y, width, height float64) {
	// Set position
	pdf.SetXY(x, y)
	
	// Draw border
	pdf.SetDrawColor(200, 200, 200)
	pdf.Rect(x, y, width, height, "D")
	
	// Image area (top 60% of cell)
	imageHeight := height * 0.6
	imageWidth := width * 0.8
	imageX := x + (width-imageWidth)/2
	imageY := y + 5
	
	// Add cover image if available
	if book.CoverImagePath != "" && fileExists(book.CoverImagePath) {
		// Try to add image, but continue if it fails
		pdf.ImageOptions(book.CoverImagePath, imageX, imageY, imageWidth, imageHeight, 
			false, gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: false}, 0, "")
	} else {
		// Draw placeholder rectangle
		pdf.SetFillColor(240, 240, 240)
		pdf.Rect(imageX, imageY, imageWidth, imageHeight, "F")
		pdf.SetXY(imageX, imageY+imageHeight/2-2)
		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(imageWidth, 4, "No Cover", "0", 0, "C", false, 0, "")
	}
	
	// Text area (bottom 40% of cell)
	textY := y + imageHeight + 10
	
	pdf.SetXY(x+2, textY)
	pdf.SetFont("Arial", "B", 8)
	
	// Title (truncated to fit)
	title := truncateString(book.Title, 25)
	pdf.CellFormat(width-4, 3, title, "0", 1, "L", false, 0, "")
	
	// Author
	pdf.SetX(x+2)
	pdf.SetFont("Arial", "", 7)
	author := truncateString(book.Author, 30)
	pdf.CellFormat(width-4, 3, author, "0", 1, "L", false, 0, "")
	
	// Date read
	pdf.SetX(x+2)
	dateStr := book.DateRead.Format("1/2/2006")
	pdf.CellFormat(width-4, 3, dateStr, "0", 1, "L", false, 0, "")
	
	// Lexile level (if available)
	if book.LexileLevel != "" {
		pdf.SetX(x+2)
		pdf.CellFormat(width-4, 3, "Lexile: "+book.LexileLevel, "0", 1, "L", false, 0, "")
	}
	
	// ISBN (if available)
	if book.ISBN != "" {
		pdf.SetX(x+2)
		isbn := truncateString(book.ISBN, 15)
		pdf.CellFormat(width-4, 3, "ISBN: "+isbn, "0", 1, "L", false, 0, "")
	}
	
	// Partial comment (if available)
	if book.IsPartial && book.PartialComment != "" {
		pdf.SetX(x+2)
		comment := truncateString(book.PartialComment, 30)
		pdf.CellFormat(width-4, 3, "Note: "+comment, "0", 1, "L", false, 0, "")
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