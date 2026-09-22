package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	MaxFailedAttempts = 5
	LockoutMinutes    = 15
)

func HashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(salt + password))
	return hex.EncodeToString(hash[:])
}

func GenerateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func ParsePasswordHash(stored string) (salt, hash string) {
	parts := strings.SplitN(stored, "$", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", stored
}

func VerifyPassword(password, storedHash string) bool {
	salt, expectedHash := ParsePasswordHash(storedHash)
	if salt == "" {
		return false
	}
	actualHash := HashPassword(password, salt)
	return actualHash == expectedHash
}

func FormatPasswordHash(password string) string {
	salt := GenerateSalt()
	hash := HashPassword(password, salt)
	return fmt.Sprintf("%s$%s", salt, hash)
}
