package middleware

import (
	"fmt"
	"log"
	"net/http"
	"rethink/api/db"
	"strings"

	"github.com/go-redis/redis"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		var tokenString string
		authHeader := c.Request().Header.Get("Authorization")

		// Check if the Authorization header is present
		authHeader = c.Request().Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// If not found, fallback to checking the cookie
			cookie, err := c.Cookie("Authorization")
			if err != nil {
				log.Println("Error: Authorization token missing from headers and cookies")
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing authorization token"})
			}
			tokenString = cookie.Value
		}
		// Ensure "Bearer " is removed if stored in the cookie
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		// Parse and verify JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(viper.GetString("JWT_SECRET")), nil
		})

		if err != nil {
			log.Println("Error: Failed to parse JWT token -", err)
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
		}

		// Extract claims from token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Println("Error: Invalid token claims")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
		}

		userID, ok := claims["userid"].(string)
		if !ok {
			log.Println("Error: userid missing in token claims")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "user_id missing in token"})
		}

		// Retrieve the stored token from Redis
		redisClient := db.GetRedisClient()
		storedToken, err := redisClient.Get("jwt:" + userID).Result()

		if err == redis.Nil {
			log.Println("Error: Token not found in Redis")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "token expired or not found"})
		} else if err != nil {
			log.Println("Error: Redis error -", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to verify token"})
		}

		// Compare the token from Redis with the provided token
		if tokenString != storedToken {
			log.Println("Error: Token mismatch")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
		}

		// Set claims in context for next middleware/handler
		c.Set("user", claims)
		return next(c)
	}
}
