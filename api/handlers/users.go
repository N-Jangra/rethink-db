package handlers

import (
	"fmt"
	"log"
	"net/http"
	"rethink/api/db"
	"rethink/api/models"
	"rethink/api/repo"
	"time"

	r "github.com/rethinkdb/rethinkdb-go"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Register handles user registration
func Register(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {

		var user models.AppUser

		//bind request data
		if err := c.Bind(&user); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid input"})
		}

		// Generate a new UUID for Userid
		user.Userid = uuid.New().String()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		// Add user to the database
		newUser, err := uc.AddUser(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Return success response with user info and token
		return c.JSON(http.StatusCreated, echo.Map{
			"message": "User Registered successfully",
			"user": echo.Map{
				"userid":   newUser.Userid,
				"name":     newUser.Name,
				"dob":      newUser.Dob,
				"email":    newUser.Email,
				"role":     newUser.Role,
				"password": newUser.Password,
			},
		})
	}
}

// Login handles user authentication and JWT generation
func Login(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Define the request body structure
		var loginRequest struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		// Bind the incoming request body to the loginRequest struct
		if err := c.Bind(&loginRequest); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid input"})
		}

		// Fetch the user from the database by Email
		user, err := uc.GetUserByEmail(loginRequest.Email)
		if err != nil {
			fmt.Println("User not found in DB:", loginRequest.Email)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid email"})
		}

		// Set active = true in the database
		res, err := r.Table("users").
			Filter(r.Row.Field("Email").Eq(user.Email)).
			Update(map[string]interface{}{"Active": true}).
			RunWrite(uc.GetSession())

		if err != nil {
			fmt.Println("Error updating active status:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update active status"})
		}

		fmt.Printf("Update result: %+v\n", res) // Print update result

		/*if res.Replaced == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found or already active"})
		}*/

		// Generate the JWT token
		token, err := generateJWT(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		}

		// Return the success message and the generated token
		return c.JSON(http.StatusOK, map[string]string{
			"message": "login successful",
			"token":   token,
			"active":  "true",
		})
	}
}

// GetUser fetches user profile
func GetUser(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {

		claims, ok := c.Get("user").(jwt.MapClaims)
		if !ok {
			fmt.Println("Error: JWT claims conversion failed")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
		}

		fmt.Println("Extracted Claims:", claims)

		userID, exists := claims["email"].(string)
		if !exists {
			fmt.Println("Error: userid not found in claims")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token format"})
		}

		// Fetch user from database
		userData, err := uc.GetUserByEmail(userID)
		if err != nil {
			fmt.Println("Error fetching user from DB:", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "user not found"})
		}

		return c.JSON(http.StatusOK, userData)
	}
}

// GetUser fetches user profile
func GetAllEmailsHandler(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {
		emails, err := uc.GetAllEmails()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch emails"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"emails": emails,
		})
	}
}

// UpdateUser updates user profile
func UpdateUser(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract email from the URL
		email := c.Param("email")
		if email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email is required"})
		}

		// Log user ID for debugging
		fmt.Println("Updating user with email:", email)

		// Fetch user from database
		userData, err := uc.GetUserByEmail(email)
		if err != nil {
			fmt.Println("Error fetching user from DB:", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "user not found"})
		}

		// Parse request body
		var updatedUser models.AppUser
		if err := c.Bind(&updatedUser); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
		}

		//field you dont want to change
		userData.CreatedAt = updatedUser.CreatedAt
		userData.Userid = updatedUser.Userid
		userData.Dob = updatedUser.Dob
		userData.Gender = updatedUser.Gender

		//changable fields
		updatedUser.UpdatedAt = time.Now()

		// Update user in DB
		err = uc.UpdateUser(email, updatedUser)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "User updated successfully"})
	}
}

// DeleteUser removes a user from the system
func DeleteUser(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get email from URL
		email := c.Param("email")
		fmt.Println("Deleting user with email:", email)
		if email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing user ID"})
		}

		// delete user from db
		err := uc.DeleteUser(email)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "user deleted successfully"})
	}
}

// Logout invalidates the user's session
func Logout(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(jwt.MapClaims)
		if !ok {
			fmt.Println("Error: JWT claims conversion failed")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
		}

		userID, ok := claims["userid"].(string)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user ID in token"})
		}

		// Remove the token from Redis
		redisClient := db.GetRedisClient()
		err := redisClient.Del("jwt:" + userID).Err()
		if err != nil {
			log.Println("Error deleting token from Redis:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to log out"})
		}

		log.Println("Token removed from Redis for user:", userID)

		return c.JSON(http.StatusOK, map[string]string{"message": "logout successful"})
	}
}
