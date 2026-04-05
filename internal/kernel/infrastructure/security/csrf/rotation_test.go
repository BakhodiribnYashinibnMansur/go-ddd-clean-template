package csrf

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotationManager(t *testing.T) {
	secret := make([]byte, 32)
	rand.Read(secret)

	gen := NewGenerator(Config{
		Secret:     secret,
		Expiration: 1 * time.Hour,
	})
	store := NewMemoryStore()
	rm := NewRotationManager(gen, store)

	t.Run("RotateOnLogin", func(t *testing.T) {
		t.Run("success_new_session", func(t *testing.T) {
			ctx := t.Context()
			newSessionID := "new-session-123"

			token, err := rm.RotateOnLogin(ctx, "", newSessionID)

			require.NoError(t, err)
			require.NotNil(t, token)
			assert.Equal(t, newSessionID, token.SessionID)

			// Verify token is stored
			storedHash, _, err := store.Get(ctx, newSessionID)
			require.NoError(t, err)
			assert.Equal(t, token.Hash, storedHash)
		})

		t.Run("success_invalidates_old_session", func(t *testing.T) {
			ctx := t.Context()
			oldSessionID := "old-session-123"
			newSessionID := "new-session-456"

			// Create old token
			oldToken, _ := gen.GenerateToken(oldSessionID)
			store.Set(ctx, oldSessionID, oldToken.Hash, 1*time.Hour)

			// Rotate
			newToken, err := rm.RotateOnLogin(ctx, oldSessionID, newSessionID)

			require.NoError(t, err)
			require.NotNil(t, newToken)

			// Old token should be deleted
			_, _, err = store.Get(ctx, oldSessionID)
			assert.Error(t, err)

			// New token should exist
			storedHash, _, err := store.Get(ctx, newSessionID)
			require.NoError(t, err)
			assert.Equal(t, newToken.Hash, storedHash)
		})
	})

	t.Run("RotateOnPasswordChange", func(t *testing.T) {
		t.Run("success_generates_new_token", func(t *testing.T) {
			ctx := t.Context()
			sessionID := "session-123"

			// Create initial token
			oldToken, _ := gen.GenerateToken(sessionID)
			store.Set(ctx, sessionID, oldToken.Hash, 1*time.Hour)

			// Rotate on password change
			newToken, err := rm.RotateOnPasswordChange(ctx, sessionID)

			require.NoError(t, err)
			require.NotNil(t, newToken)

			// Tokens should be different
			assert.NotEqual(t, oldToken.Value, newToken.Value)
			assert.NotEqual(t, oldToken.Hash, newToken.Hash)

			// New token should be stored
			storedHash, _, err := store.Get(ctx, sessionID)
			require.NoError(t, err)
			assert.Equal(t, newToken.Hash, storedHash)
		})
	})

	t.Run("RotateOnPrivilegeChange", func(t *testing.T) {
		t.Run("success_rotates_token", func(t *testing.T) {
			ctx := t.Context()
			sessionID := "session-123"

			// Create initial token
			oldToken, _ := gen.GenerateToken(sessionID)
			store.Set(ctx, sessionID, oldToken.Hash, 1*time.Hour)

			// Rotate on privilege change
			newToken, err := rm.RotateOnPrivilegeChange(ctx, sessionID)

			require.NoError(t, err)
			require.NotNil(t, newToken)

			// Verify rotation
			storedHash, _, err := store.Get(ctx, sessionID)
			require.NoError(t, err)
			assert.Equal(t, newToken.Hash, storedHash)
		})
	})

	t.Run("RotateOnRefresh", func(t *testing.T) {
		t.Run("success_rotates_token", func(t *testing.T) {
			ctx := t.Context()
			sessionID := "session-123"

			// Create initial token
			oldToken, _ := gen.GenerateToken(sessionID)
			store.Set(ctx, sessionID, oldToken.Hash, 1*time.Hour)

			// Rotate on refresh
			newToken, err := rm.RotateOnRefresh(ctx, sessionID)

			require.NoError(t, err)
			require.NotNil(t, newToken)

			// Verify rotation
			storedHash, _, err := store.Get(ctx, sessionID)
			require.NoError(t, err)
			assert.Equal(t, newToken.Hash, storedHash)
		})
	})

	t.Run("InvalidateOnLogout", func(t *testing.T) {
		t.Run("success_deletes_token", func(t *testing.T) {
			ctx := t.Context()
			sessionID := "session-123"

			// Create token
			token, _ := gen.GenerateToken(sessionID)
			store.Set(ctx, sessionID, token.Hash, 1*time.Hour)

			// Invalidate on logout
			err := rm.InvalidateOnLogout(ctx, sessionID)

			require.NoError(t, err)

			// Token should be deleted
			_, _, err = store.Get(ctx, sessionID)
			assert.Error(t, err)
		})
	})

	t.Run("SessionFixationProtection", func(t *testing.T) {
		t.Run("prevents_session_fixation_attack", func(t *testing.T) {
			ctx := t.Context()

			// Attacker creates a session and gets CSRF token
			attackerSessionID := "attacker-session"
			attackerToken, _ := gen.GenerateToken(attackerSessionID)
			store.Set(ctx, attackerSessionID, attackerToken.Hash, 1*time.Hour)

			// Victim logs in with new session
			victimSessionID := "victim-session"
			victimToken, err := rm.RotateOnLogin(ctx, attackerSessionID, victimSessionID)

			require.NoError(t, err)

			// Attacker's old token should not work
			_, _, err = store.Get(ctx, attackerSessionID)
			assert.Error(t, err, "Old session should be invalidated")

			// Only victim's new token should work
			storedHash, _, err := store.Get(ctx, victimSessionID)
			require.NoError(t, err)
			assert.Equal(t, victimToken.Hash, storedHash)
		})
	})

	t.Run("ReplayAttackProtection", func(t *testing.T) {
		t.Run("prevents_replay_after_rotation", func(t *testing.T) {
			ctx := t.Context()
			sessionID := "session-123"

			// Create initial token
			oldToken, _ := gen.GenerateToken(sessionID)
			store.Set(ctx, sessionID, oldToken.Hash, 1*time.Hour)

			// Rotate token
			newToken, _ := rm.RotateOnPasswordChange(ctx, sessionID)

			// Try to validate old token
			storedHash, expiresAt, _ := store.Get(ctx, sessionID)
			err := gen.ValidateToken(oldToken.Value, storedHash, sessionID, expiresAt)

			// Old token should fail validation
			assert.Error(t, err, "Old token should not validate after rotation")

			// New token should validate
			err = gen.ValidateToken(newToken.Value, newToken.Hash, sessionID, newToken.ExpiresAt)
			assert.NoError(t, err, "New token should validate")
		})
	})
}
