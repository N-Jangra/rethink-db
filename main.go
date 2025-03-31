package main

import (
	"rethink/api/db"
	"rethink/api/handlers"
	"rethink/api/repo"
	"rethink/api/routes"

	"github.com/labstack/echo/v4"
)

func main() {
	dbinstance := db.InitDB()
	db.InitRedis()

	e := echo.New()
	//e.Static("/", "static")
	e.Static("/api/web", "./api/web")

	userController := repo.NewUserController(dbinstance)
	bookController := repo.NewBookController(dbinstance)

	handlers.UserRoute(e, userController)
	handlers.BooksRoute(e, bookController)

	routes.UserRoutes(e)
	routes.BookRoutes(e)

	e.GET("/", handlers.Home)

	e.Logger.Fatal(e.Start(":8090"))
}
