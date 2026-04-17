package credentials

import (
	"golang.org/x/crypto/bcrypt"
)


// HashPassword generates a hash of the given password using a default cost.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}


// CheckPasswordHash compares a plain text password with a hashed password, returns true if they are equal.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}