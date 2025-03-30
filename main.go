package main

import (
	"rethink/api/db"
	"rethink/api/handlers"
	"rethink/api/repo"
	"rethink/api/routes"

	"github.com/labstack/echo/v4"
	gorethink "github.com/rethinkdb/rethinkdb-go"
)

func main() {
	db.InitDB()
	db.InitRedis()

	e := echo.New()
	//e.Static("/", "static")
	e.Static("/api/web", "./api/web")

	userController := repo.NewUserController(&gorethink.Session{})
	bookController := repo.NewBookController(&gorethink.Session{})

	handlers.UserRoute(e, userController)
	handlers.BooksRoute(e, bookController)

	routes.UserRoutes(e)
	routes.BookRoutes(e)

	e.GET("/", handlers.Home)

	e.Logger.Fatal(e.Start(":8090"))
}
