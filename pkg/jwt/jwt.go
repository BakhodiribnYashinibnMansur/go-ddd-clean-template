package jwt

import (
	"crypto/rsa"
	"fmt"
	"strings"

	jwtgo "github.com/golang-jwt/jwt/v4"
)

// ParseRSAPrivateKey parses an RSA private key from a PEM string.
func ParseRSAPrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	keyStr = strings.ReplaceAll(keyStr, PEMNewlineEscaped, PEMNewline)
	key, err := jwtgo.ParseRSAPrivateKeyFromPEM([]byte(keyStr))
	if err != nil {
		return nil, fmt.Errorf(ErrMsgParsePrivateKey, err)
	}
	return key, nil
}

// ParseRSAPublicKey parses an RSA public key from a PEM string.
func ParseRSAPublicKey(keyStr string) (*rsa.PublicKey, error) {
	keyStr = strings.ReplaceAll(keyStr, PEMNewlineEscaped, PEMNewline)
	key, err := jwtgo.ParseRSAPublicKeyFromPEM([]byte(keyStr))
	if err != nil {
		return nil, fmt.Errorf(ErrMsgParsePublicKey, err)
	}
	return key, nil
}
