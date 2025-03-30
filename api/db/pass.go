package db

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err) // Handle this properly in production
	}
	return string(hashedPassword)
}

func CheckHashedPassword(plainPassword, plainPasswordPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(plainPassword), []byte(plainPassword))
	return err == nil
}
