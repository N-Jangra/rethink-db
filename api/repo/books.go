package repo

import (
	"errors"
	"log"
	"rethink/api/models"

	r "github.com/rethinkdb/rethinkdb-go"
)

// bookController struct handles database interactions for books
type BookController struct {
	Session *r.Session
}

// NewbookController initializes the bookController with a GORM DB instance
func NewBookController(Session *r.Session) *BookController {
	return &BookController{Session: Session}

}

// Getbooks retrieves all books from the database
func (bc *BookController) GetBooks() ([]models.Books, error) {

	log.Println("Fetching Books from DB...")
	var books []models.Books

	cursor, err := r.Table("books").Run(bc.Session)
	if err != nil {
		log.Println("Error Fetching Books:", err)
		return nil, err
	}
	defer cursor.Close()

	err = cursor.All(&books)
	if err != nil {
		log.Println("Error parsing books:", err)
		return nil, err
	}

	log.Println("Books Retrieved:", books)
	return books, nil
}

// Getbook retrieves a single book by ID
func (bc *BookController) GetBook(BookID int) (*models.Books, error) {

	var book models.Books
	cursor, err := r.Table("books").Filter(r.Row.Field("BookID").Eq(BookID)).Run(bc.Session)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	if cursor.IsNil() {
		return nil, errors.New("book not found")
	}

	err = cursor.One(&book)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

// Createbook adds a new book to the database
func (bc *BookController) CreateBook(book *models.Books) (models.AppUser, error) {
	res, err := r.Table("books").Insert(book).RunWrite(bc.Session)
	if err != nil {
		log.Println("Error inserting book:", err)
		return models.AppUser{}, err
	}
	if res.Inserted == 0 {
		return models.AppUser{}, errors.New("failed to insert book")
	}
	return models.AppUser{}, nil
}

// Updatebook updates an existing book in the database
func (bc *BookController) UpdateBook(BookID int, updatedbook models.Books) error {

	res, err := r.Table("books").
		Filter(r.Row.Field("BookID").Eq(BookID)).
		Update(updatedbook, r.UpdateOpts{NonAtomic: true}).
		RunWrite(bc.Session)

	if err != nil {
		return err
	}

	if res.Replaced == 0 && res.Updated == 0 {
		return errors.New("no book found or no changes made")
	}

	return nil
}

// Deletebook removes an book from the database
func (bc *BookController) DeleteBook(BookID int) error {
	res, err := r.Table("books").Filter(r.Row.Field("BookID").Eq(BookID)).Delete().RunWrite(bc.Session)
	if err != nil {
		return err
	}

	if res.Deleted == 0 {
		return errors.New("book not found or already deleted")
	}

	return nil
}
