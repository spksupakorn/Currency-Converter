package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/argon2"
)

func GenerateSalt(length int) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(salt), nil
}

func HashPasswordArgon2(password, salt string) string {
	hash := argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(hash) + ":" + salt
}
func VerifyPasswordArgon2(password, hashed string) bool {
	parts := strings.Split(hashed, ":")
	if len(parts) != 2 {
		return false
	}
	hash := argon2.IDKey([]byte(password), []byte(parts[1]), 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(hash) == parts[0]
}
