package jwt

import (
	"crypto/rsa"
	"fmt"
	"strings"

	jwtgo "github.com/golang-jwt/jwt/v4"
)

// ParseRSAPrivateKey parses an RSA private key from a PEM string.
func ParseRSAPrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	if keyStr == "" {
		return nil, fmt.Errorf("jwt - ParseRSAPrivateKey: key string is empty")
	}

	// Clean string from potential surrounding quotes and spaces
	cleanKey := strings.Trim(keyStr, "\" '`\t\n\r")
	// Handle literal \n and \\n escapes commonly found in env files
	cleanKey = strings.ReplaceAll(cleanKey, "\\\\n", "\n")
	cleanKey = strings.ReplaceAll(cleanKey, "\\n", "\n")

	key, err := jwtgo.ParseRSAPrivateKeyFromPEM([]byte(cleanKey))
	if err != nil {
		// Log a hint about the key format to help debugging without exposing the whole key
		truncated := cleanKey
		if len(truncated) > 20 {
			truncated = truncated[:20] + "..."
		}
		return nil, fmt.Errorf("%s (key start: %q, length: %d): %w", "jwt - ParseRSAPrivateKey", truncated, len(cleanKey), err)
	}
	return key, nil
}

// ParseRSAPublicKey parses an RSA public key from a PEM string.
func ParseRSAPublicKey(keyStr string) (*rsa.PublicKey, error) {
	if keyStr == "" {
		return nil, fmt.Errorf("jwt - ParseRSAPublicKey: key string is empty")
	}

	// Clean string from potential surrounding quotes and spaces
	cleanKey := strings.Trim(keyStr, "\" '`\t\n\r")
	// Handle literal \n and \\n escapes commonly found in env files
	cleanKey = strings.ReplaceAll(cleanKey, "\\\\n", "\n")
	cleanKey = strings.ReplaceAll(cleanKey, "\\n", "\n")

	key, err := jwtgo.ParseRSAPublicKeyFromPEM([]byte(cleanKey))
	if err != nil {
		truncated := cleanKey
		if len(truncated) > 20 {
			truncated = truncated[:20] + "..."
		}
		return nil, fmt.Errorf("%s (key start: %q, length: %d): %w", "jwt - ParseRSAPublicKey", truncated, len(cleanKey), err)
	}
	return key, nil
}
