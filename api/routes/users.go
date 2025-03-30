package routes

import (
	"rethink/api/db"
	"rethink/api/handlers"
	"rethink/api/middleware"
	"rethink/api/repo"

	"github.com/labstack/echo/v4"
)

// UserRoutes initializes user-related API endpoints
func UserRoutes(e *echo.Echo) {
	dbInstance := db.InitDB()
	uc := repo.NewUserController(dbInstance)

	e.POST("/register", handlers.Register(uc))                                      //register
	e.POST("/login", handlers.Login(uc))                                            //login
	e.GET("/logout", handlers.Logout(uc), middleware.AuthMiddleware)                //logout
	e.GET("/profile", handlers.GetUser(uc), middleware.AuthMiddleware)              //read user
	e.GET("/profiles", handlers.GetAllEmailsHandler(uc))                            //read all user
	e.PUT("/profile/:email", handlers.UpdateUser(uc), middleware.AuthMiddleware)    //update user
	e.DELETE("/profile/:email", handlers.DeleteUser(uc), middleware.AuthMiddleware) //delete user
}
