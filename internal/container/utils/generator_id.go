package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random ID with 6 bytes (12 hex characters)
func GenerateID() (string, error) {
	bytes := make([]byte, 6)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
