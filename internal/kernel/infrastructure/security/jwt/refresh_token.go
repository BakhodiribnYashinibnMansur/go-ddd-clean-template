package jwt

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrRefreshTokenInvalid is returned when a refresh token cannot be
	// parsed or fails format checks.
	ErrRefreshTokenInvalid = errors.New("invalid refresh token")
	// ErrRefreshTokenExpired is returned by callers holding the stored
	// expiration alongside the hash; this package does not self-check expiry.
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	// ErrHashMismatch is returned when the stored HMAC does not match the
	// secret presented in the token.
	ErrHashMismatch = errors.New("refresh token hash mismatch")
	// ErrPepperTooShort is returned by NewRefreshHasher when the pepper is
	// below the minimum safe length (32 bytes / 256 bits).
	ErrPepperTooShort = errors.New("refresh hasher: pepper must be at least 32 bytes")
)

// RefreshHasher derives and verifies refresh-token hashes using HMAC-SHA256
// with a server-side pepper. A single instance is safe for concurrent use.
//
// We pick HMAC (not PBKDF2/bcrypt/argon2) because the input is already 256
// bits of crypto/rand output: slow hashing buys nothing, while HMAC with a
// secret pepper blocks offline attacks against a DB dump.
type RefreshHasher struct {
	pepper []byte
}

// NewRefreshHasher constructs a RefreshHasher. The pepper must be at least
// 32 raw bytes. Rotate the pepper to invalidate all outstanding refresh
// tokens (e.g. after a credential leak).
func NewRefreshHasher(pepper []byte) (*RefreshHasher, error) {
	if len(pepper) < 32 {
		return nil, ErrPepperTooShort
	}
	// Defensive copy — caller may mutate the passed slice.
	cp := make([]byte, len(pepper))
	copy(cp, pepper)
	return &RefreshHasher{pepper: cp}, nil
}

// Hash returns the base64-URL (no padding) HMAC-SHA256 digest of the token
// secret, bound to the token ID via a domain-separation prefix.
func (h *RefreshHasher) Hash(secret, tokenID string) string {
	mac := hmac.New(sha256.New, h.pepper)
	// Length-prefixing prevents collisions between (secret="ab", id="cd")
	// and (secret="abc", id="d"); we use unambiguous "|" separators too.
	mac.Write([]byte(refreshHashDomain))
	mac.Write([]byte{'|'})
	mac.Write([]byte(tokenID))
	mac.Write([]byte{'|'})
	mac.Write([]byte(secret))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Verify compares a recomputed hash of (secret, tokenID) against storedHash
// using a constant-time comparison.
func (h *RefreshHasher) Verify(secret, tokenID, storedHash string) bool {
	computed := h.Hash(secret, tokenID)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(storedHash)) == 1
}

// RefreshToken is the in-memory representation of a refresh token. It is
// never directly serialized to the wire — use String() for the client
// representation and Hashed for DB persistence.
type RefreshToken struct {
	ID        string    // Random token ID (also used as HMAC-binding input).
	Secret    string    // Random secret; only populated during generation or parsing.
	Hashed    string    // HMAC of (secret, ID); safe to store server-side.
	UserID    string    // Subject (user ID).
	SessionID string    // Session ID.
	IssuedAt  time.Time // Issued-at timestamp (UTC).
	ExpiresAt time.Time // Expiration timestamp (UTC).
	ClientID  string    // Client/device identifier.
}

// GenerateRefreshToken creates a new refresh token, sets its Hashed field via
// the supplied hasher, and returns it to the caller. The caller must persist
// token.Hashed (and the expiration) on the session and return token.String()
// to the client.
func GenerateRefreshToken(
	hasher *RefreshHasher,
	userID, sessionID, clientID string,
	expiresIn time.Duration,
) (*RefreshToken, error) {
	if hasher == nil {
		return nil, fmt.Errorf("jwt.GenerateRefreshToken: hasher is nil")
	}

	tokenID, err := randomURLSafeString(refreshTokenIDBytes)
	if err != nil {
		return nil, fmt.Errorf("jwt.GenerateRefreshToken: id: %w", err)
	}
	secret, err := randomURLSafeString(refreshSecretBytes)
	if err != nil {
		return nil, fmt.Errorf("jwt.GenerateRefreshToken: secret: %w", err)
	}

	now := time.Now().UTC()
	t := &RefreshToken{
		ID:        tokenID,
		Secret:    secret,
		UserID:    userID,
		SessionID: sessionID,
		ClientID:  clientID,
		IssuedAt:  now,
		ExpiresAt: now.Add(expiresIn),
	}
	t.Hashed = hasher.Hash(secret, tokenID)
	return t, nil
}

// String returns the client-facing token: rft_v1.<sid>.<id>.<secret>.
func (t *RefreshToken) String() string {
	return RefreshTokenPrefix + RefreshTokenVersion + "." + t.SessionID + "." + t.ID + "." + t.Secret
}

// IsExpired is true if the token's ExpiresAt is in the past (UTC).
func (t *RefreshToken) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// ParseRefreshToken splits the wire format into its components. It performs
// ONLY structural validation — the caller must then call
// RefreshHasher.Verify(secret, id, storedHash) to authenticate the token.
func ParseRefreshToken(tokenString string) (*RefreshToken, error) {
	if !strings.HasPrefix(tokenString, RefreshTokenPrefix) {
		return nil, fmt.Errorf("%w: invalid prefix", ErrRefreshTokenInvalid)
	}
	parts := strings.SplitN(tokenString[len(RefreshTokenPrefix):], ".", 4)
	if len(parts) != 4 {
		return nil, fmt.Errorf("%w: invalid format", ErrRefreshTokenInvalid)
	}
	if parts[0] != RefreshTokenVersion {
		return nil, fmt.Errorf("%w: unsupported version", ErrRefreshTokenInvalid)
	}
	if parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return nil, fmt.Errorf("%w: empty component", ErrRefreshTokenInvalid)
	}
	return &RefreshToken{
		SessionID: parts[1],
		ID:        parts[2],
		Secret:    parts[3],
	}, nil
}

// randomURLSafeString returns a crypto/rand-derived string of nBytes raw
// entropy, encoded using base64 URL without padding.
func randomURLSafeString(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("jwt.rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
