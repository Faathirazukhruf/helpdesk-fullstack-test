package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword mengenkripsi password plain-text menjadi hash bcrypt yang aman.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword membandingkan password plain-text dengan hash yang tersimpan di database.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
