package utils

import "golang.org/x/crypto/bcrypt"

func GeneratePasswordHash(s string) (string, error) {
	saltedBytes := []byte(s)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	hash := string(hashedBytes[:])
	return hash, nil
}
func Compare(hash string, s string) error {
	incoming := []byte(s)
	existing := []byte(hash)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}
func ComparePassword(hashedPass, pass string) error {
	err := Compare(hashedPass, pass)
	if err != nil {
		return err
	}
	return nil
}
