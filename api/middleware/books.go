package middleware

import (
	"log"
	"net/http"
	"rethink/api/repo"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// CheckAccess is a middleware to verify user authorization
func CheckAccess(db *repo.UserController, requiredPermission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Retrieve claims from the context
			userClaims, ok := c.Get("user").(jwt.MapClaims)
			if !ok {
				log.Println("JWT Token missing from context")
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "token not found in context"})
			}

			// Extract email (assuming email is used as a unique identifier)
			email, ok := userClaims["email"].(string)
			if !ok {
				log.Println("Email missing in token claims")
				return c.JSON(http.StatusForbidden, echo.Map{"error": "email missing in token"})
			}

			// Retrieve user role from RethinkDB
			role, err := db.GetUserRoleByEmail(email)
			if err != nil {
				log.Println("Error fetching role from DB:", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch role"})
			}

			// Debugging: Log user role
			log.Println("User Role:", role)

			// Check role permissions in DB
			hasPermission, err := db.HasPermission(role, requiredPermission)
			if err != nil {
				log.Println("Error checking permissions:", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to verify permissions"})
			}

			if !hasPermission {
				log.Println("Access Denied for Role:", role, "Required:", requiredPermission)
				return c.JSON(http.StatusForbidden, echo.Map{"error": "access denied"})
			}

			return next(c)
		}
	}
}
