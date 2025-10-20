package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/stretchr/testify/assert"
)

// Test data for various edge cases
func createTestBooks() []*BookForPDF {
	return []*BookForPDF{
		// Normal book
		{
			Title:       "The Great Gatsby",
			Author:      "F. Scott Fitzgerald",
			ISBN:        "9780743273565",
			LexileLevel: "1070L",
			DateRead:    time.Now(),
			CoverURL:    "https://example.com/cover1.jpg",
			IsPartial:   false,
		},
		// Extra long title and author
		{
			Title:       "The Incredibly Long and Extraordinarily Detailed Title of a Book That Should Test Text Wrapping Functionality",
			Author:      "An Author With an Extremely Long Name That Should Also Test The Text Wrapping Capability",
			ISBN:        "9780123456789",
			LexileLevel: "850L",
			DateRead:    time.Now(),
			CoverURL:    "",
			IsPartial:   true,
		},
		// No image, short text
		{
			Title:       "1984",
			Author:      "George Orwell",
			ISBN:        "9780451524935",
			LexileLevel: "950L",
			DateRead:    time.Now(),
			CoverURL:    "",
			IsPartial:   false,
		},
		// Medium length text
		{
			Title:       "To Kill a Mockingbird: A Classic American Novel",
			Author:      "Harper Lee",
			ISBN:        "9780061120084",
			LexileLevel: "870L",
			DateRead:    time.Now(),
			CoverURL:    "https://example.com/cover4.jpg",
			IsPartial:   false,
		},
	}
}

// Create 32+ books for full page testing
func createLargeBookSet() []*BookForPDF {
	books := make([]*BookForPDF, 35) // Test with 35 books to ensure pagination
	
	titles := []string{
		"Short Title",
		"This is a Much Longer Title That Will Test Text Wrapping",
		"Medium Length Book Title",
		"Extremely Long Title That Goes On And On And Should Definitely Wrap Across Multiple Lines",
		"Another Book",
	}
	
	authors := []string{
		"Short Author",
		"An Author With A Very Long Name That Should Wrap",
		"Medium Name",
		"Extremely Long Author Name That Goes On Forever And Ever",
		"Simple Name",
	}
	
	for i := 0; i < 35; i++ {
		books[i] = &BookForPDF{
			Title:       titles[i%len(titles)] + " " + fmt.Sprintf("#%d", i+1),
			Author:      authors[i%len(authors)],
			ISBN:        fmt.Sprintf("978012345%04d", i),
			LexileLevel: fmt.Sprintf("%dL", 700+i*10),
			DateRead:    time.Now().AddDate(0, 0, -i),
			CoverURL:    "", // Mix of images and no images
			IsPartial:   i%3 == 0, // Every third book is partial
		}
		
		// Add cover URL for some books
		if i%2 == 0 {
			books[i].CoverURL = fmt.Sprintf("https://example.com/cover%d.jpg", i)
		}
	}
	
	return books
}

func TestWrapText(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	
	tests := []struct {
		name      string
		text      string
		maxWidth  float64
		expectMin int // minimum expected lines
		expectMax int // maximum expected lines
	}{
		{
			name:      "Short text",
			text:      "Short title",
			maxWidth:  100.0,
			expectMin: 1,
			expectMax: 1,
		},
		{
			name:      "Long text that should wrap",
			text:      "This is a very long title that should definitely wrap across multiple lines when constrained",
			maxWidth:  50.0,
			expectMin: 2,
			expectMax: 5,
		},
		{
			name:      "Single long word",
			text:      "Supercalifragilisticexpialidocious",
			maxWidth:  30.0,
			expectMin: 1,
			expectMax: 1, // Should truncate with ...
		},
		{
			name:      "Empty text",
			text:      "",
			maxWidth:  100.0,
			expectMin: 1,
			expectMax: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := wrapText(pdf, tt.text, tt.maxWidth)
			assert.GreaterOrEqual(t, len(lines), tt.expectMin, "Should have at least %d lines", tt.expectMin)
			assert.LessOrEqual(t, len(lines), tt.expectMax, "Should have at most %d lines", tt.expectMax)
			
			// Verify no line is too wide
			for i, line := range lines {
				width := pdf.GetStringWidth(line)
				assert.LessOrEqual(t, width, tt.maxWidth, "Line %d '%s' is too wide: %.2f > %.2f", i, line, width, tt.maxWidth)
			}
		})
	}
}

func TestCalculateRowHeight(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	tests := []struct {
		name        string
		book        *BookForPDF
		columnWidth float64
		minHeight   float64
	}{
		{
			name: "Short title and author",
			book: &BookForPDF{
				Title:  "1984",
				Author: "George Orwell",
			},
			columnWidth: 45.0,
			minHeight:   30.0,
		},
		{
			name: "Very long title and author",
			book: &BookForPDF{
				Title:  "The Incredibly Long and Extraordinarily Detailed Title of a Book That Should Test Text Wrapping Functionality",
				Author: "An Author With an Extremely Long Name That Should Also Test The Text Wrapping Capability",
			},
			columnWidth: 45.0,
			minHeight:   35.0, // Should be taller due to text wrapping
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			height := calculateRowHeight(pdf, tt.book, tt.columnWidth)
			assert.GreaterOrEqual(t, height, tt.minHeight, "Row height should be at least %.1f", tt.minHeight)
			assert.LessOrEqual(t, height, 100.0, "Row height should be reasonable (< 100mm)")
		})
	}
}

func TestCreateTestPDF(t *testing.T) {
	// Test with edge case books
	t.Run("Edge cases PDF", func(t *testing.T) {
		books := createTestBooks()
		pdf := createTestPDF("Test Child", "Edge Cases", books)
		
		// Save to temp file
		tempFile := "/tmp/test_edge_cases.pdf"
		err := pdf.OutputFileAndClose(tempFile)
		assert.NoError(t, err, "Should create PDF without error")
		
		// Verify file exists and has content
		info, err := os.Stat(tempFile)
		assert.NoError(t, err, "PDF file should exist")
		assert.Greater(t, info.Size(), int64(1000), "PDF should have substantial content")
		
		// Clean up
		os.Remove(tempFile)
	})
	
	// Test with 35 books (full page + pagination)
	t.Run("Large book set PDF", func(t *testing.T) {
		books := createLargeBookSet()
		pdf := createTestPDF("Test Child", "Full Page Test", books)
		
		// Save to temp file
		tempFile := "/tmp/test_large_set.pdf"
		err := pdf.OutputFileAndClose(tempFile)
		assert.NoError(t, err, "Should create PDF without error")
		
		// Verify file exists and has content
		info, err := os.Stat(tempFile)
		assert.NoError(t, err, "PDF file should exist")
		assert.Greater(t, info.Size(), int64(3000), "PDF with 35 books should have substantial content")
		
		// Clean up
		os.Remove(tempFile)
	})
}

// Helper function to create test PDF without database dependencies
func createTestPDF(childName, title string, books []*BookForPDF) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	
	// Header
	header := fmt.Sprintf("%s - %s", childName, title)
	pdf.Cell(0, 10, header)
	pdf.Ln(15)
	
	// Page dimensions
	pageWidth, pageHeight := pdf.GetPageSize()
	leftMargin, topMargin, rightMargin, bottomMargin := pdf.GetMargins()
	usableWidth := pageWidth - leftMargin - rightMargin
	
	// Calculate layout for 4-column display with dynamic row heights
	columnsPerRow := 4
	columnSpacing := 3.0
	columnWidth := (usableWidth - (columnSpacing * float64(columnsPerRow-1))) / float64(columnsPerRow)
	
	// Start drawing books
	currentY := topMargin + 25
	
	for i, book := range books {
		// Calculate column position (0, 1, 2, or 3)
		columnIndex := i % columnsPerRow
		columnX := leftMargin + float64(columnIndex)*(columnWidth+columnSpacing)
		
		// Calculate row height needed for this book
		rowHeight := calculateRowHeight(pdf, book, columnWidth)
		
		// Check if we need a new page
		if currentY+rowHeight > pageHeight-bottomMargin-30 {
			pdf.AddPage()
			pdf.SetFont("Arial", "B", 16)
			pdf.Cell(0, 10, header)
			pdf.Ln(15)
			currentY = topMargin + 25
		}
		
		drawBookColumn(pdf, book, columnX, currentY, columnWidth, rowHeight)
		
		// Move to next row after 4th column
		if columnIndex == 3 {
			currentY += rowHeight + 2
		}
	}
	
	return pdf
}