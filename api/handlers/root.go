package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Home(c echo.Context) error {
	return c.Redirect(http.StatusSeeOther, "/login")
}
