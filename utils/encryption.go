package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func EncryptPassword(p string) ([]byte, error) {
	bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
	return bs, err
}

func CheckPassword(userPassword []byte, p string) bool {
	err := bcrypt.CompareHashAndPassword(userPassword, []byte(p))
	if err != nil {
		return false
	}
	return true
}
