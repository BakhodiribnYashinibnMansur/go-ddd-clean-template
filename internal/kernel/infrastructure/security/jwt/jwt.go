// Package jwt provides asymmetric (RS256) access-token signing/verification
// and a stateless refresh-token format with server-side HMAC hashing.
//
// Security properties:
//   - Access tokens are RS256 only; the parser rejects any other algorithm.
//   - Audience, issuer, expiry, and issued-at are enforced on every parse
//     (with a configurable leeway for clock skew).
//   - Refresh tokens carry only a random ID and secret on the wire; the server
//     stores HMAC-SHA256(pepper, secret) and compares it in constant time.
//   - PEM parsing errors never echo key material.
package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"

	jwtgo "github.com/golang-jwt/jwt/v5"
)

// ErrEmptyKey is returned when a nil or empty PEM buffer is supplied.
var ErrEmptyKey = errors.New("jwt: key is empty")

// ParseRSAPrivateKey parses an RSA private key from a PEM byte buffer.
//
// The caller is responsible for any I/O sanitization (quote stripping,
// \n-escape expansion). This function intentionally does none — it is a
// pure crypto primitive. See config.cleanConfigStrings for the loader-side
// normalization used by this project.
func ParseRSAPrivateKey(pem []byte) (*rsa.PrivateKey, error) {
	if len(pem) == 0 {
		return nil, ErrEmptyKey
	}
	key, err := jwtgo.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		// Deliberately omit any bytes of the key from the error — we only
		// surface the length as a debugging hint.
		return nil, fmt.Errorf("jwt: parse RSA private key (len=%d): %w", len(pem), err)
	}
	return key, nil
}

// ParseRSAPublicKey parses an RSA public key from a PEM byte buffer.
// Same caveats as ParseRSAPrivateKey apply.
func ParseRSAPublicKey(pem []byte) (*rsa.PublicKey, error) {
	if len(pem) == 0 {
		return nil, ErrEmptyKey
	}
	key, err := jwtgo.ParseRSAPublicKeyFromPEM(pem)
	if err != nil {
		return nil, fmt.Errorf("jwt: parse RSA public key (len=%d): %w", len(pem), err)
	}
	return key, nil
}
