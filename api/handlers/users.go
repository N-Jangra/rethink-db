package handlers

import (
	"fmt"
	"log"
	"net/http"
	"rethink/api/db"
	"rethink/api/models"
	"rethink/api/repo"
	"strings"
	"time"

	r "github.com/rethinkdb/rethinkdb-go"
	"github.com/spf13/viper"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Register handles user registration
func Register(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {

		var user models.AppUser

		// Manually extract form values
		user.Name = c.FormValue("name")
		user.Email = c.FormValue("email")
		user.Password = c.FormValue("password")
		user.Role = c.FormValue("role")
		user.Gender = c.FormValue("sex")
		user.Phone = c.FormValue("phone")
		user.Details = c.FormValue("details")

		// Parse the date of birth (DOB) field
		dobStr := c.FormValue("dob") // Get dob as a string from the form
		if dobStr != "" {
			parsedDob, err := time.Parse("2006-01-02", dobStr) // Expecting YYYY-MM-DD format
			if err != nil {
				fmt.Println("Error parsing DOB:", err) // Debugging output
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date format. Use YYYY-MM-DD"})
			}
			user.Dob = parsedDob
		} else {
			fmt.Println("DOB not provided, skipping...")
		}

		// Generate a new UUID for Userid
		user.Userid = uuid.New().String()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		// Add user to the database
		_, err := uc.AddUser(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Return success response with user info and token
		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title":   "Register",
			"Message": "User created successfully!",
		})
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles user authentication and JWT generation
func Login(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Define the request body structure
		loginRequest := new(LoginRequest)

		// Bind the incoming form data
		if err := c.Bind(&loginRequest); err != nil {
			fmt.Println("JSON bind failed, trying form values...")
			// If JSON bind fails, try manually getting form values
			loginRequest.Email = c.FormValue("email")
			loginRequest.Password = c.FormValue("password")
		}
		fmt.Printf("Attempting login with email: %s\n", loginRequest.Email)

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

		fmt.Printf("Update result: %+v\n", res)

		// Generate the JWT token
		token, err := generateJWT(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		}

		// Store token in session
		c.SetCookie(&http.Cookie{
			Name:     "Authorization",
			Value:    "Bearer " + token,
			HttpOnly: true,
			Path:     "/",
		})

		return c.Redirect(http.StatusSeeOther, "/boks")
	}
}

// GetUser fetches user profile
func GetUser(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Extract JWT token from cookies
		cookie, err := c.Cookie("Authorization")
		if err != nil {
			fmt.Println("Error: JWT token not found in cookies")
			return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
				"Title": "User Details",
				"Error": "Unauthorized: No token found",
			})
		}

		// Remove "Bearer " prefix from the token
		tokenString := strings.TrimPrefix(cookie.Value, "Bearer ")
		fmt.Println("Extracted Token:", tokenString)

		// Parse the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the correct signing method is used
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(getJWTSecret()), nil
		})

		// Check if the token is valid
		if err != nil || !token.Valid {
			fmt.Println("Error: Invalid JWT token -", err)
			return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
				"Title": "User Details",
				"Error": "Invalid token",
			})
		}

		// Extract claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Println("Error: JWT claims conversion failed")
			return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
				"Title": "User Details",
				"Error": "Invalid token claims",
			})
		}

		// Extract email
		email, exists := claims["email"].(string)
		if !exists {
			fmt.Println("Error: email not found in claims")
			return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
				"Title": "User Details",
				"Error": "Invalid token format",
			})
		}
		fmt.Println("Extracted Email:", email)

		// Fetch user from database
		userData, err := uc.GetUserByEmail(email)
		if err != nil {
			fmt.Println("Error fetching user from DB:", err)
			return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
				"Title": "User Details",
				"Error": "User not found",
			})
		}

		// Store user data in context for templates
		c.Set("userData", userData)

		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title": "User Details",
			"User": map[string]interface{}{
				"UserID":   userData.Userid,
				"Name":     userData.Name,
				"Dob":      userData.Dob.Format("2006-01-02"),
				"Email":    userData.Email,
				"Role":     userData.Role,
				"Gender":   userData.Gender,
				"Phone":    userData.Phone,
				"Details":  userData.Details,
				"JoinedAt": userData.CreatedAt.Format("2006-01-02 15:04:05"),
			},
		})
	}
}

// UpdateUser updates user profile
func UpdateUser(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract email from the URL
		fmt.Println("Form values received:", c.Request().Form)
		email := c.FormValue("email")
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

		// Manually bind form data
		var updatedUser models.AppUser
		updatedUser.Name = c.FormValue("name")
		updatedUser.Email = email
		updatedUser.Gender = c.FormValue("sex")
		updatedUser.Details = c.FormValue("details")
		updatedUser.Phone = c.FormValue("phone")
		updatedUser.Password = c.FormValue("password")

		//field you dont want to change
		userData.CreatedAt = updatedUser.CreatedAt
		userData.Userid = updatedUser.Userid
		userData.Dob = updatedUser.Dob
		userData.Role = updatedUser.Role

		//changable fields
		updatedUser.UpdatedAt = time.Now()

		// Update user in DB
		err = uc.UpdateUser(email, updatedUser)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.Render(http.StatusOK, "layout.html", map[string]interface{}{
			"Title":   "Update User",
			"Message": "User updated successfully",
		})
	}
}

// DeleteUser removes a user from the system
func DeleteUser(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Get email from URL
		email := c.FormValue("email")
		fmt.Println("Deleting user with email:", email)

		if email == "" {
			fmt.Println("Error: Email is missing in form submission")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing user ID"})
		}

		// delete user from db
		err := uc.DeleteUser(email)
		if err != nil {
			fmt.Println("Error deleting user:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.Redirect(http.StatusSeeOther, "/login")
	}
}

// Logout invalidates the user's session
func Logout(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Retrieve the token from cookies instead of the Authorization header
		cookie, err := c.Cookie("Authorization")
		if err != nil {
			log.Println("Error: Authorization cookie missing")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authorization token missing"})
		}

		// Extract the token value
		tokenString := strings.TrimPrefix(cookie.Value, "Bearer ")

		// Validate and parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(viper.GetString("JWT_SECRET")), nil
		})

		if err != nil {
			log.Println("Error parsing token:", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		// Extract claims from token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
		}

		userID, ok := claims["userid"].(string)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user ID in token"})
		}
		Email := claims["email"].(string)

		// Fetch the user from the database by Email
		user, err := uc.GetUserByEmail(Email)
		if err != nil {
			fmt.Println("User not found in DB:", Email)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid email"})
		}

		// Set active = true in the database
		res, err := r.Table("users").
			Filter(r.Row.Field("Email").Eq(user.Email)).
			Update(map[string]interface{}{"Active": false}).
			RunWrite(uc.GetSession())

		if err != nil {
			fmt.Println("Error updating active status:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update active status"})
		}

		fmt.Printf("Update result: %+v\n", res) // Print update result

		// Remove the token from Redis
		redisClient := db.GetRedisClient()
		err = redisClient.Del("jwt:" + userID).Err()
		if err != nil {
			log.Println("Error deleting token from Redis:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to log out"})
		}

		// Clear the cookie in the response
		c.SetCookie(&http.Cookie{
			Name:     "Authorization",
			Value:    "",
			Expires:  time.Unix(0, 0),
			Path:     "/",
			HttpOnly: true,
		})

		log.Println("Token removed from Redis for user:", userID)

		return c.Redirect(http.StatusFound, "/login")
	}
}
