package jwt

import (
	"crypto/rsa"
	"fmt"

	jwtgo "github.com/dgrijalva/jwt-go"
)

// ParseRSAPrivateKey parses an RSA private key from a PEM string.
func ParseRSAPrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	key, err := jwtgo.ParseRSAPrivateKeyFromPEM([]byte(keyStr))
	if err != nil {
		return nil, fmt.Errorf("jwt - ParseRSAPrivateKey - ParseRSAPrivateKeyFromPEM: %w", err)
	}
	return key, nil
}

// ParseRSAPublicKey parses an RSA public key from a PEM string.
func ParseRSAPublicKey(keyStr string) (*rsa.PublicKey, error) {
	key, err := jwtgo.ParseRSAPublicKeyFromPEM([]byte(keyStr))
	if err != nil {
		return nil, fmt.Errorf("jwt - ParseRSAPublicKey - ParseRSAPublicKeyFromPEM: %w", err)
	}
	return key, nil
}
