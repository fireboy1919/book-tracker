package main

import (
	"fmt"
	"github.com/booktracker/backend/config"
)

func main() {
	config.InitDatabase()
	db := config.GetDB()
	
	// Check for books with invalid child_id
	var invalidBooks int
	db.Raw("SELECT COUNT(*) FROM books WHERE child_id NOT IN (SELECT id FROM children)").Scan(&invalidBooks)
	fmt.Printf("Books with invalid child_id: %d\n", invalidBooks)
	
	// Check for books with invalid shared_book_id
	var invalidSharedBooks int
	db.Raw("SELECT COUNT(*) FROM books WHERE shared_book_id IS NOT NULL AND shared_book_id NOT IN (SELECT id FROM shared_books)").Scan(&invalidSharedBooks)
	fmt.Printf("Books with invalid shared_book_id: %d\n", invalidSharedBooks)
	
	// Check total books and children
	var totalBooks, totalChildren int
	db.Raw("SELECT COUNT(*) FROM books").Scan(&totalBooks)
	db.Raw("SELECT COUNT(*) FROM children").Scan(&totalChildren)
	fmt.Printf("Total books: %d, Total children: %d\n", totalBooks, totalChildren)
}