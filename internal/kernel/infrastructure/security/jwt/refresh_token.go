package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

const (
	refreshTokenVersion = "v1"
	tokenPrefix         = "rft_"
	saltSize            = 16
	keyLength           = 32
	iterations          = 10000
)

var (
	ErrRefreshTokenInvalid = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrHashMismatch        = errors.New("invalid refresh token: hash mismatch")
)

// RefreshToken represents a refresh token with its components
type RefreshToken struct {
	ID        string    `json:"jti"`       // Unique token ID
	Secret    string    `json:"-"`         // Secret part (only for initial generation)
	Hashed    string    `json:"hashed"`    // Hashed version for storage
	UserID    string    `json:"sub"`       // Subject (user ID)
	SessionID string    `json:"sid"`       // Session ID
	IssuedAt  time.Time `json:"iat"`       // Issued at
	ExpiresAt time.Time `json:"exp"`       // Expiration time
	ClientID  string    `json:"client_id"` // Client/device identifier
}

// GenerateRefreshToken creates a new refresh token
func GenerateRefreshToken(userID, sessionID, clientID string, expiresIn time.Duration) (*RefreshToken, error) {
	// Generate a unique token ID
	tokenID, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token ID: %w", err)
	}

	// Generate a random secret
	secret, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token secret: %w", err)
	}

	now := time.Now()
	token := &RefreshToken{
		ID:        tokenID,
		Secret:    secret,
		UserID:    userID,
		SessionID: sessionID,
		ClientID:  clientID,
		IssuedAt:  now,
		ExpiresAt: now.Add(expiresIn),
	}

	// Generate the hashed version for storage
	token.Hashed = hashToken(secret, tokenID)

	return token, nil
}

// String returns the full token string in format: rft_v1.<session_id>.<id>.<secret>
func (t *RefreshToken) String() string {
	return fmt.Sprintf("%s%s.%s.%s.%s", tokenPrefix, refreshTokenVersion, t.SessionID, t.ID, t.Secret)
}

// Verify checks if the token is valid and matches the stored hash
func (t *RefreshToken) Verify(storedHash string) bool {
	if t.IsExpired() {
		return false
	}
	return verifyHash(t.Secret, t.ID, storedHash)
}

// IsExpired checks if the token has expired
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// ParseRefreshToken parses a refresh token string
func ParseRefreshToken(tokenString string) (*RefreshToken, error) {
	// Check prefix and version
	if !strings.HasPrefix(tokenString, tokenPrefix) {
		return nil, fmt.Errorf("%w: invalid prefix", ErrRefreshTokenInvalid)
	}

	// Remove prefix and split into parts using "." separator
	parts := strings.SplitN(tokenString[len(tokenPrefix):], ".", 4)
	if len(parts) < 4 {
		return nil, fmt.Errorf("%w: invalid format", ErrRefreshTokenInvalid)
	}

	version := parts[0]
	if version != refreshTokenVersion {
		return nil, fmt.Errorf("%w: unsupported version", ErrRefreshTokenInvalid)
	}

	return &RefreshToken{
		SessionID: parts[1],
		ID:        parts[2],
		Secret:    parts[3],
	}, nil
}

// VerifyRefreshToken verifies a refresh token against its stored hash.
// Note: This only verifies the hash, not expiry. Expiry should be checked
// separately using the stored expiration time from the database.
func VerifyRefreshToken(tokenString, hashedSecret string) (*RefreshToken, error) {
	token, err := ParseRefreshToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if !verifyHash(token.Secret, token.ID, hashedSecret) {
		return nil, ErrHashMismatch
	}

	// Add the hash to the token for consistency
	token.Hashed = hashedSecret
	return token, nil
}

// hashToken generates a secure hash of the token secret
func hashToken(secret, salt string) string {
	hash := pbkdf2.Key(
		[]byte(secret),
		[]byte(salt),
		iterations,
		keyLength,
		sha256.New,
	)
	return base64.URLEncoding.EncodeToString(hash)
}

// verifyHash checks if the secret matches the stored hash
func verifyHash(secret, salt, storedHash string) bool {
	return hashToken(secret, salt) == storedHash
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("jwt.refresh_token.rand: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
