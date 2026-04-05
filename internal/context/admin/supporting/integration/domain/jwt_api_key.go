package domain

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
)

// ErrInvalidJWTAPIKey is returned when a JWT API key fails length validation.
var ErrInvalidJWTAPIKey = errors.New("invalid JWT API key: must be >=32 chars")

// minJWTAPIKeyLength is the minimum acceptable length for a JWT API key plaintext.
const minJWTAPIKeyLength = 32

// JWTAPIKey is a high-entropy secret presented in the X-API-Key header to
// identify which integration should issue/validate a JWT. The plaintext is
// never stored server-side — only its HMAC-SHA256 hash (keyed with a
// server-side pepper). Construct one from a 32+ byte base64url random.
type JWTAPIKey struct {
	plaintext string
}

// NewJWTAPIKey validates and creates a JWTAPIKey. Requires a minimum
// plaintext length of 32 characters.
func NewJWTAPIKey(s string) (JWTAPIKey, error) {
	if len(s) < minJWTAPIKeyLength {
		return JWTAPIKey{}, ErrInvalidJWTAPIKey
	}
	return JWTAPIKey{plaintext: s}, nil
}

// String returns a redacted placeholder safe for logs and error messages.
func (k JWTAPIKey) String() string { return "[REDACTED]" }

// Reveal returns the raw plaintext. Intended for single-use egress such as
// handing the key back to the operator immediately after provisioning.
func (k JWTAPIKey) Reveal() string { return k.plaintext }

// IsZero reports whether this JWTAPIKey is the zero value.
func (k JWTAPIKey) IsZero() bool { return k.plaintext == "" }

// Hash returns HMAC-SHA256(pepper, plaintext) — the server-side storage form.
func (k JWTAPIKey) Hash(pepper []byte) []byte {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(k.plaintext))
	return mac.Sum(nil)
}

// VerifyJWTAPIKeyHash recomputes HMAC-SHA256(pepper, plaintext) and compares
// against storedHash in constant time.
func VerifyJWTAPIKeyHash(plaintext string, pepper []byte, storedHash []byte) bool {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(plaintext))
	computed := mac.Sum(nil)
	return subtle.ConstantTimeCompare(computed, storedHash) == 1
}

// GenerateJWTAPIKey returns a new JWTAPIKey backed by 32 random bytes encoded
// as base64url (RawURLEncoding), yielding a 43-character plaintext.
func GenerateJWTAPIKey() (JWTAPIKey, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return JWTAPIKey{}, fmt.Errorf("generate jwt api key: %w", err)
	}
	return NewJWTAPIKey(base64.RawURLEncoding.EncodeToString(buf))
}
