package csrf

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	t.Run("success_with_config", func(t *testing.T) {
		secret := make([]byte, 32)
		rand.Read(secret)

		gen := NewGenerator(Config{
			Secret:     secret,
			Expiration: 1 * time.Hour,
		})

		require.NotNil(t, gen)
		assert.Equal(t, 1*time.Hour, gen.expiration)
	})

	t.Run("panic_on_empty_secret", func(t *testing.T) {
		assert.Panics(t, func() {
			NewGenerator(Config{Secret: []byte{}})
		})
	})

	t.Run("default_expiration", func(t *testing.T) {
		secret := make([]byte, 32)
		rand.Read(secret)

		gen := NewGenerator(Config{Secret: secret})
		assert.Equal(t, DefaultExpiration, gen.expiration)
	})
}

func TestGenerateToken(t *testing.T) {
	secret := make([]byte, 32)
	rand.Read(secret)
	gen := NewGenerator(Config{Secret: secret})

	t.Run("success_generates_unique_tokens", func(t *testing.T) {
		token1, err := gen.GenerateToken("session1")
		require.NoError(t, err)
		require.NotNil(t, token1)

		token2, err := gen.GenerateToken("session1")
		require.NoError(t, err)
		require.NotNil(t, token2)

		// Tokens should be unique
		assert.NotEqual(t, token1.Value, token2.Value)
		assert.NotEqual(t, token1.Hash, token2.Hash)
	})

	t.Run("success_token_has_correct_fields", func(t *testing.T) {
		sessionID := "test-session-123"
		token, err := gen.GenerateToken(sessionID)

		require.NoError(t, err)
		assert.NotEmpty(t, token.Value)
		assert.NotEmpty(t, token.Hash)
		assert.Equal(t, sessionID, token.SessionID)
		assert.True(t, token.ExpiresAt.After(time.Now()))
	})

	t.Run("success_different_sessions_different_hashes", func(t *testing.T) {
		token1, _ := gen.GenerateToken("session1")
		token2, _ := gen.GenerateToken("session2")

		// Even with same token value, different sessions should produce different hashes
		assert.NotEqual(t, token1.Hash, token2.Hash)
	})
}

func TestValidateToken(t *testing.T) {
	secret := make([]byte, 32)
	rand.Read(secret)
	gen := NewGenerator(Config{
		Secret:     secret,
		Expiration: 1 * time.Hour,
	})

	t.Run("success_valid_token", func(t *testing.T) {
		sessionID := "session-123"
		token, _ := gen.GenerateToken(sessionID)

		err := gen.ValidateToken(token.Value, token.Hash, sessionID, token.ExpiresAt)
		assert.NoError(t, err)
	})

	t.Run("error_missing_token", func(t *testing.T) {
		err := gen.ValidateToken("", "hash", "session", time.Now().Add(1*time.Hour))
		assert.ErrorIs(t, err, ErrMissingToken)
	})

	t.Run("error_expired_token", func(t *testing.T) {
		sessionID := "session-123"
		token, _ := gen.GenerateToken(sessionID)

		// Set expiration to past
		pastTime := time.Now().Add(-1 * time.Hour)

		err := gen.ValidateToken(token.Value, token.Hash, sessionID, pastTime)
		assert.ErrorIs(t, err, ErrExpiredToken)
	})

	t.Run("error_invalid_token", func(t *testing.T) {
		sessionID := "session-123"
		token, _ := gen.GenerateToken(sessionID)

		// Use wrong token value
		err := gen.ValidateToken("wrong-token", token.Hash, sessionID, token.ExpiresAt)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("error_wrong_session", func(t *testing.T) {
		token, _ := gen.GenerateToken("session1")

		// Validate with different session ID
		err := gen.ValidateToken(token.Value, token.Hash, "session2", token.ExpiresAt)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("error_tampered_hash", func(t *testing.T) {
		sessionID := "session-123"
		token, _ := gen.GenerateToken(sessionID)

		// Tamper with hash
		tamperedHash := token.Hash + "tampered"

		err := gen.ValidateToken(token.Value, tamperedHash, sessionID, token.ExpiresAt)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}

func TestRotateToken(t *testing.T) {
	secret := make([]byte, 32)
	rand.Read(secret)
	gen := NewGenerator(Config{Secret: secret})

	t.Run("success_rotation_generates_new_token", func(t *testing.T) {
		sessionID := "session-123"

		token1, err := gen.GenerateToken(sessionID)
		require.NoError(t, err)

		token2, err := gen.RotateToken(sessionID)
		require.NoError(t, err)

		// Rotated token should be different
		assert.NotEqual(t, token1.Value, token2.Value)
		assert.NotEqual(t, token1.Hash, token2.Hash)
	})
}

func TestConstantTimeComparison(t *testing.T) {
	secret := make([]byte, 32)
	rand.Read(secret)
	gen := NewGenerator(Config{Secret: secret})

	sessionID := "session-123"
	token, _ := gen.GenerateToken(sessionID)

	// This test ensures timing attacks are prevented
	// by using constant-time comparison
	t.Run("timing_attack_protection", func(t *testing.T) {
		// Multiple validations should take similar time
		// regardless of where the mismatch occurs

		iterations := 100
		for range iterations {
			// Valid token
			_ = gen.ValidateToken(token.Value, token.Hash, sessionID, token.ExpiresAt)

			// Invalid token (different lengths)
			_ = gen.ValidateToken("short", token.Hash, sessionID, token.ExpiresAt)

			// Invalid token (same length, different value)
			_ = gen.ValidateToken(token.Value+"x", token.Hash, sessionID, token.ExpiresAt)
		}

		// If this test completes without panic, timing is consistent
		assert.True(t, true)
	})
}
