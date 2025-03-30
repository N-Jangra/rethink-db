package handlers

import (
	"fmt"
	"log"
	"net/http"
	"rethink/api/models"
	"rethink/api/repo"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	r "github.com/rethinkdb/rethinkdb-go"
)

// GetbooksHandler retrieves all books
func Getbooks(bc *repo.BookController) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Println("Handling /books GET request...")
		books, err := bc.GetBooks()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, books)
	}
}

// GetbookHandler retrieves a specific book by ID
func Getbook(bc *repo.BookController) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Extract BookID from request URL and convert it to an integer
		BookID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book ID"})
		}

		book, err := bc.GetBook(BookID)
		if err != nil {
			if err == r.ErrEmptyResult {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "book not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
		}
		return c.JSON(http.StatusOK, book)
	}
}

// CreatebookHandler creates a new book
func Createbook(bc *repo.BookController) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Retrieve user from context
		userClaims, exists := c.Get("user").(jwt.MapClaims)
		if !exists {
			fmt.Println("Error: User context is missing")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
		}

		// Extract user ID from claims
		userID, exists := userClaims["userid"].(string)
		if !exists {
			fmt.Println("Error: userid not found in claims")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token format"})
		}

		// Debugging: Output the user ID from token claims
		fmt.Println("Authenticated User ID:", userID)

		// Create a new book
		var book models.Books
		if err := c.Bind(&book); err != nil {
			fmt.Println("Binding Error:", err) // Print the actual error
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book data"})
		}

		// Find max bookid and increment it by 1, handling empty table case
		var maxID int
		cursor, err := r.Table("books").Max("BookID").Default(map[string]int{"BookID": 0}).Pluck("BookID").Run(bc.Session)
		if err != nil {
			fmt.Println("Error fetching max BookID:", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to get max book ID"})
		}
		defer cursor.Close()

		var result map[string]int
		if cursor.Next(&result) {
			maxID = result["BookID"]
			book.BookID = maxID + 1
		} else {
			book.BookID = 1 // If no books exist, start from 1
		}

		// Check for cursor errors
		if err := cursor.Err(); err != nil {
			fmt.Println("Cursor Error:", err)
			book.BookID = 1
		}

		// Debug: Log the book data after binding
		fmt.Printf("Received Book Data: %+v\n", book)

		// Set the createdBy and updatedBy fields using the userID from JWT claims
		book.CreatedBy = userID
		book.UpdatedBy = userID
		book.CreatedAt = time.Now()
		book.UpdatedAt = time.Now()

		// Store the book in database
		_, err = bc.CreateBook(&book)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create book"})
		}

		// Return success response with book info
		return c.JSON(http.StatusCreated, echo.Map{
			"message": "Book Created successfully",
			"book":    book,
		})
	}
}

// UpdatebookHandler updates an book
func Updatebook(bc *repo.BookController) echo.HandlerFunc {
	return func(c echo.Context) error {
		BookID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book ID"})
		}

		// Fetch the existing book to preserve certain fields
		existingBook, err := bc.GetBook(BookID)
		if err != nil {
			if err == r.ErrEmptyResult {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "book not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch book"})
		}

		var updatedbook models.Books
		if err := c.Bind(&updatedbook); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid input"})
		}

		// Update only the specified fields, keeping other fields unchanged
		existingBook.Title = updatedbook.Title
		existingBook.Description = updatedbook.Description
		existingBook.UpdatedBy = updatedbook.UpdatedBy
		existingBook.UpdatedAt = time.Now() // Set current timestamp

		err = bc.UpdateBook(BookID, *existingBook)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "book updated successfully"})
	}
}

// DeletebookHandler deletes an book
func Deletebook(bc *repo.BookController) echo.HandlerFunc {
	return func(c echo.Context) error {
		BookID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book ID"})
		}

		err = bc.DeleteBook(BookID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "book deleted successfully"})
	}
}
