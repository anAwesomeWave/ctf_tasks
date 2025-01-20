package util

import "golang.org/x/crypto/bcrypt"

func GetHashPassword(password string) (string, error) {
	bytePassword := []byte(password)
	hash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func IsHashEqualPassword(pHash, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(pHash), []byte(plainPassword))
	return err == nil
}
