package routes

import (
	"rethink/api/db"
	"rethink/api/handlers"
	"rethink/api/middleware"
	"rethink/api/repo"

	"github.com/labstack/echo/v4"
)

func BookRoutes(e *echo.Echo) {

	dbInstance := db.InitDB()
	bc := repo.NewBookController(dbInstance)
	uc := repo.NewUserController(dbInstance)

	e.GET("/books", handlers.Getbooks(bc), middleware.AuthMiddleware, middleware.CheckAccess(uc, "book_read"))
	e.GET("/books/:id", handlers.Getbook(bc), middleware.AuthMiddleware, middleware.CheckAccess(uc, "book_read"))
	e.POST("/books", handlers.Createbook(bc), middleware.AuthMiddleware, middleware.CheckAccess(uc, "book_create"))
	e.PUT("/books/:id", handlers.Updatebook(bc), middleware.AuthMiddleware, middleware.CheckAccess(uc, "book_update"))
	e.DELETE("/books/:id", handlers.Deletebook(bc), middleware.AuthMiddleware, middleware.CheckAccess(uc, "book_delete"))
}
