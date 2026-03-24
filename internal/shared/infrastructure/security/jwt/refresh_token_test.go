package jwt

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken("user-1", "session-1", "client-1", 7*24*time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token.ID)
	assert.NotEmpty(t, token.Secret)
	assert.NotEmpty(t, token.Hashed)
	assert.Equal(t, "user-1", token.UserID)
	assert.Equal(t, "session-1", token.SessionID)
	assert.Equal(t, "client-1", token.ClientID)
	assert.False(t, token.IsExpired())
	assert.True(t, token.ExpiresAt.After(time.Now()))
}

func TestRefreshToken_String(t *testing.T) {
	token, err := GenerateRefreshToken("user-1", "sess-1", "client-1", time.Hour)
	require.NoError(t, err)

	tokenStr := token.String()
	assert.True(t, strings.HasPrefix(tokenStr, "rft_v1."))
	assert.Contains(t, tokenStr, "sess-1")
}

func TestParseRefreshToken_Valid(t *testing.T) {
	original, err := GenerateRefreshToken("user-1", "sess-1", "client-1", time.Hour)
	require.NoError(t, err)

	parsed, err := ParseRefreshToken(original.String())
	require.NoError(t, err)
	assert.Equal(t, original.ID, parsed.ID)
	assert.Equal(t, original.SessionID, parsed.SessionID)
	assert.Equal(t, original.Secret, parsed.Secret)
}

func TestParseRefreshToken_InvalidPrefix(t *testing.T) {
	_, err := ParseRefreshToken("invalid_v1.sess.id.secret")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
}

func TestParseRefreshToken_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "too_few_parts",
			token: "rft_v1.only",
		},
		{
			name:  "wrong_version",
			token: "rft_v999.sess.id.secret",
		},
		{
			name:  "empty_string",
			token: "",
		},
		{
			name:  "no_prefix",
			token: "v1.sess.id.secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRefreshToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestRefreshToken_Verify(t *testing.T) {
	token, err := GenerateRefreshToken("user-1", "sess-1", "client-1", time.Hour)
	require.NoError(t, err)

	// Verify with correct hash
	assert.True(t, token.Verify(token.Hashed))

	// Verify with wrong hash
	assert.False(t, token.Verify("wrong-hash"))
}

func TestRefreshToken_IsExpired(t *testing.T) {
	// Generate a token that is already expired
	token, err := GenerateRefreshToken("user-1", "sess-1", "client-1", -1*time.Hour)
	require.NoError(t, err)
	assert.True(t, token.IsExpired())

	// Verify fails for expired token
	assert.False(t, token.Verify(token.Hashed))
}

func TestVerifyRefreshToken_Success(t *testing.T) {
	original, err := GenerateRefreshToken("user-1", "sess-1", "client-1", time.Hour)
	require.NoError(t, err)

	verified, err := VerifyRefreshToken(original.String(), original.Hashed)
	require.NoError(t, err)
	assert.Equal(t, original.ID, verified.ID)
	assert.Equal(t, original.Hashed, verified.Hashed)
}

func TestVerifyRefreshToken_WrongHash(t *testing.T) {
	original, err := GenerateRefreshToken("user-1", "sess-1", "client-1", time.Hour)
	require.NoError(t, err)

	_, err = VerifyRefreshToken(original.String(), "wrong-hash-value")
	assert.ErrorIs(t, err, ErrHashMismatch)
}

func TestVerifyRefreshToken_InvalidToken(t *testing.T) {
	_, err := VerifyRefreshToken("not-a-valid-token", "some-hash")
	assert.Error(t, err)
}

func TestHashToken_Deterministic(t *testing.T) {
	hash1 := hashToken("secret", "salt")
	hash2 := hashToken("secret", "salt")
	assert.Equal(t, hash1, hash2)

	// Different inputs produce different hashes
	hash3 := hashToken("different-secret", "salt")
	assert.NotEqual(t, hash1, hash3)
}

func TestGenerateRandomString(t *testing.T) {
	s1, err := generateRandomString(32)
	require.NoError(t, err)
	assert.NotEmpty(t, s1)

	s2, err := generateRandomString(32)
	require.NoError(t, err)
	assert.NotEqual(t, s1, s2, "two random strings should not be equal")
}
