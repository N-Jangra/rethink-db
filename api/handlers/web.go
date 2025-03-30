package handlers

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"rethink/api/db"
	"rethink/api/middleware"
	"rethink/api/repo"
	"strconv"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer is a custom renderer for Echo using html/template
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func loadTemplates(contentFile string) *TemplateRenderer {
	templates := template.Must(template.New("layout.html").ParseFiles("api/web/layout.html", contentFile))
	return &TemplateRenderer{templates: templates}
}

func loadTemplates2(contentFile string) *TemplateRenderer {
	templates := template.Must(template.New("layout2.html").ParseFiles("api/web/layout2.html", contentFile))
	return &TemplateRenderer{templates: templates}
}

// LoginRoute defines the /login route
func UserRoute(e *echo.Echo, uc *repo.UserController) {

	//load login page
	e.GET("/login", func(c echo.Context) error {
		renderer := loadTemplates("api/web/login.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "Login",
		})
	})

	e.POST("/login", Login(uc))

	//load register page
	e.GET("/register", func(c echo.Context) error {
		renderer := loadTemplates("api/web/register.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "Register",
		})
	})

	e.POST("/register", Register(uc))

	//load details page
	e.GET("/user/details", func(c echo.Context) error {
		handler := GetUser(repo.NewUserController(uc.GetSession()))

		// Call the handler function to get user details
		err := handler(c)
		if err != nil {
			return err
		}

		// Get user data from response
		userData := c.Get("userData")
		if userData == nil {
			fmt.Println("Error: No user data found in context")
			return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
				"Title": "User Details : ",
				"Error": "No user data found",
			})
		}

		// Render the template with user data
		renderer := loadTemplates("api/web/userdetails.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "User Details : ",
			"User":  userData,
		})
	})

	//load update page
	e.GET("/user/update", func(c echo.Context) error {
		renderer := loadTemplates("api/web/userupdate.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "Update User : ",
		})
	})

	e.POST("/user/update", UpdateUser(uc))

	//load delete page
	e.GET("/user/delete", func(c echo.Context) error {
		renderer := loadTemplates("api/web/userdelete.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "Delete User Account",
		})
	})

	e.POST("/user/delete", DeleteUser(uc))

	e.GET("/user/logout", func(c echo.Context) error {
		renderer := loadTemplates("api/web/userlogout.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "Logout Confirmation",
		})
	})

	e.POST("/user/logout", Logout(uc))

}

func BooksRoute(e *echo.Echo, bc *repo.BookController) {

	//load books page
	e.GET("/boks", func(c echo.Context) error {
		renderer := loadTemplates("api/web/boks.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "Books Management",
		})
	})

	// load all books
	e.GET("/books/all", func(c echo.Context) error {
		// Fetch books from the database
		bc := repo.NewBookController(db.DB)
		books, err := bc.GetBooks()
		if err != nil {
			return c.Render(http.StatusOK, "layout2.html", map[string]interface{}{
				"Title": "All Books",
				"Error": "Failed to retrieve books",
			})
		}

		// Render bookall.html with book data
		renderer := loadTemplates2("api/web/bookall.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout2.html", map[string]interface{}{
			"Title": "All Books",
			"Books": books,
		})
	})

	//load a specific book
	e.GET("/books/id", func(c echo.Context) error {
		// Extract the "id" parameter from the query string if it exists
		bookIDParam := c.QueryParam("id")

		// Initialize the data map to pass to the template
		data := map[string]interface{}{
			"Title": "Get Book by ID : ",
		}

		if bookIDParam != "" {
			// Try to convert the id from the query string to an integer
			bookID, err := strconv.Atoi(bookIDParam)
			if err != nil {
				// Return an error if the ID is invalid
				data["Error"] = "Invalid book ID"
				renderer := loadTemplates("api/web/bookbyid.html")
				e.Renderer = renderer
				return c.Render(http.StatusOK, "layout.html", data)
			}

			// Fetch the book from the database
			bc := repo.NewBookController(db.DB)
			book, err := bc.GetBook(bookID)
			if err != nil {
				// Return an error if the book is not found
				data["Error"] = "Book not found"
			} else {
				// If the book is found, add it to the data map
				data["Book"] = book
			}
		}

		// Load and render the template with the data
		renderer := loadTemplates("api/web/bookbyid.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", data)
	})

	//create a new book
	e.GET("/books/create", func(c echo.Context) error {
		renderer := loadTemplates("api/web/bookcreate.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "Create a New Book : ",
		})
	})
	e.POST("/books/create", middleware.AuthMiddleware(Createbook(bc)))

	//update a book
	e.GET("/books/update", func(c echo.Context) error {
		renderer := loadTemplates("api/web/bookupdate.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "Update Book : ",
		})
	})
	e.POST("/books/update", middleware.AuthMiddleware(Updatebook(bc)))

	//delete a book
	e.GET("/books/delete", func(c echo.Context) error {
		renderer := loadTemplates("api/web/bookdelete.html")
		e.Renderer = renderer
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "Delete Book : ",
		})
	})
	e.POST("/books/delete", Deletebook(bc))

}
