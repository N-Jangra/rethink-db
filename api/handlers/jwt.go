package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"rethink/api/db"
	"rethink/api/models"
	"rethink/api/repo"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// Secret key (should be from environment variables)
func getJWTSecret() string {
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	secret := viper.GetString("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET not found in config file")
	}

	return secret
}

// Generate JWT Token (Private Function)
func generateJWT(user *models.AppUser) (string, error) {
	secret := getJWTSecret()
	claims := &jwt.MapClaims{
		"userid":    user.Userid,
		"email":     user.Email,
		"name":      user.Name,
		"role":      user.Role,
		"expiresAt": time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("Error signing JWT:", err)
		return "", err
	}

	// Store token in Redis
	redisClient := db.GetRedisClient()
	err = redisClient.Set("jwt:"+user.Userid, tokenString, time.Hour*24).Err() // Store for 24 hours

	if err != nil {
		log.Println("Error storing token in Redis:", err)
		//return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store token"})
	} else {
		log.Println("Token stored successfully.")
	}

	return tokenString, nil
}

// GenerateJWTHandler generates a JWT token for a user
func GenerateJWTHandler(uc *repo.UserController) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Get user ID from query parameters
		userID := c.QueryParam("userid")
		if userID == "" {
			// Log missing userid
			fmt.Println("Error: userid is missing in the query")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "userid is required"})
		}

		// Fetch user details
		user, err := uc.GetUserByEmail(userID)
		if err != nil {
			// Log if the user is not found
			fmt.Println("Error: user not found,", err)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}

		// Generate a JWT token for the user
		token, err := generateJWT(user)
		if err != nil {
			// Log if token generation fails
			fmt.Println("Error generating JWT:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		}

		// Return the JWT token in the response
		return c.JSON(http.StatusOK, map[string]string{"token": token})
	}
}

// ValidateJWT parses and validates the JWT token
func ValidateJWT(tokenString string) (map[string]interface{}, error) {
	secret := getJWTSecret()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		log.Println("JWT Validation Error:", err)
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Println("Decoded JWT Claims:", claims) // Debugging
		return claims, nil
	}

	log.Println("Invalid Token Received")
	return nil, errors.New("invalid token")
}

// Middleware to check JWT authentication
func CheckJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token format"})
		}

		// Validate JWT
		_, err := ValidateJWT(tokenString)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		return next(c)
	}
}

// GetClaims decodes the JWT token and returns the claims
func GetClaims(token string) (jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		// Validate signing method and return the secret key
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// extracts jwt token from authorization header
func GetJWTFromHeader(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	token := strings.Split(authHeader, " ")[1]

	return token
}
