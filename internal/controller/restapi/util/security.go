package util

import (
	"github.com/google/uuid"
)

// GenerateToken generates a new random token using UUID.
func GenerateToken() string {
	return uuid.New().String()
}
