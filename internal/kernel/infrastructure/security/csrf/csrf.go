package csrf

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidToken is returned when CSRF token validation fails
	ErrInvalidToken = errors.New("invalid CSRF token")
	// ErrExpiredToken is returned when CSRF token has expired
	ErrExpiredToken = errors.New("CSRF token has expired")
	// ErrMissingToken is returned when CSRF token is missing
	ErrMissingToken = errors.New("CSRF token is missing")
)

const (
	// TokenLength is the length of the random token in bytes
	TokenLength = 32
	// DefaultExpiration is the default token expiration time
	DefaultExpiration = 24 * time.Hour
)

// Token represents a CSRF token with metadata
type Token struct {
	Value     string    // Plain token value (sent to client)
	Hash      string    // HMAC hash of the token (stored server-side)
	ExpiresAt time.Time // Token expiration time
	SessionID string    // Associated session ID
}

// Config holds CSRF configuration
type Config struct {
	Secret     []byte        // Secret key for HMAC
	Expiration time.Duration // Token expiration duration
}

// Generator handles CSRF token generation and validation
type Generator struct {
	secret     []byte
	expiration time.Duration
}

// NewGenerator creates a new CSRF token generator
func NewGenerator(config Config) *Generator {
	if len(config.Secret) == 0 {
		panic("CSRF secret cannot be empty")
	}

	expiration := config.Expiration
	if expiration == 0 {
		expiration = DefaultExpiration
	}

	return &Generator{
		secret:     config.Secret,
		expiration: expiration,
	}
}

// GenerateToken creates a new cryptographically secure CSRF token
func (g *Generator) GenerateToken(sessionID string) (*Token, error) {
	// Generate cryptographically random bytes
	b := make([]byte, TokenLength)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("csrf.GenerateToken.rand: %w", err)
	}

	// Encode to base64 URL-safe format
	plainToken := base64.RawURLEncoding.EncodeToString(b)

	// Create HMAC hash
	hash := g.hashToken(plainToken, sessionID)

	return &Token{
		Value:     plainToken,
		Hash:      hash,
		ExpiresAt: time.Now().Add(g.expiration),
		SessionID: sessionID,
	}, nil
}

// hashToken creates HMAC-SHA256 hash of token + sessionID
func (g *Generator) hashToken(token, sessionID string) string {
	h := hmac.New(sha256.New, g.secret)
	h.Write([]byte(token + sessionID))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// ValidateToken validates a CSRF token against stored hash
// Uses constant-time comparison to prevent timing attacks
func (g *Generator) ValidateToken(plainToken, storedHash, sessionID string, expiresAt time.Time) error {
	if plainToken == "" {
		return ErrMissingToken
	}

	// Check expiration
	if time.Now().After(expiresAt) {
		return ErrExpiredToken
	}

	// Compute hash of incoming token
	computedHash := g.hashToken(plainToken, sessionID)

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(computedHash), []byte(storedHash)) != 1 {
		return ErrInvalidToken
	}

	return nil
}

// RotateToken generates a new token for session rotation
// Should be called after login, privilege change, etc.
func (g *Generator) RotateToken(sessionID string) (*Token, error) {
	return g.GenerateToken(sessionID)
}
