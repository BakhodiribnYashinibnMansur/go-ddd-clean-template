package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestRSAKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privKey, &privKey.PublicKey
}

func TestGenerateAndParseAccessToken(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	tests := []struct {
		name      string
		userID    string
		sessionID string
		issuer    string
		audience  string
		expiresIn time.Duration
		wantErr   bool
	}{
		{
			name:      "valid_token",
			userID:    "user-123",
			sessionID: "session-456",
			issuer:    "test-issuer",
			audience:  "test-audience",
			expiresIn: 15 * time.Minute,
			wantErr:   false,
		},
		{
			name:      "empty_user_id",
			userID:    "",
			sessionID: "session-456",
			issuer:    "test-issuer",
			audience:  "test-audience",
			expiresIn: 15 * time.Minute,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStr, err := GenerateAccessToken(tt.userID, tt.sessionID, tt.issuer, tt.audience, privKey, tt.expiresIn)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, tokenStr)

			// Parse back
			claims, err := ParseAccessToken(tokenStr, pubKey, tt.issuer, tt.audience)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.sessionID, claims.SessionID)
			assert.Equal(t, TokenTypeAccess, claims.Type)
		})
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	tokenStr, err := GenerateAccessToken("user-1", "sess-1", "iss", "aud", privKey, -1*time.Minute)
	require.NoError(t, err)

	_, err = ParseAccessToken(tokenStr, pubKey, "iss", "aud")
	assert.ErrorIs(t, err, ErrAccessTokenExpired)
}

func TestParseAccessToken_WrongKey(t *testing.T) {
	privKey1, _ := generateTestRSAKeys(t)
	_, pubKey2 := generateTestRSAKeys(t) // different key pair

	tokenStr, err := GenerateAccessToken("user-1", "sess-1", "iss", "aud", privKey1, 15*time.Minute)
	require.NoError(t, err)

	_, err = ParseAccessToken(tokenStr, pubKey2, "iss", "aud")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

func TestParseAccessToken_WrongIssuer(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	tokenStr, err := GenerateAccessToken("user-1", "sess-1", "correct-issuer", "aud", privKey, 15*time.Minute)
	require.NoError(t, err)

	_, err = ParseAccessToken(tokenStr, pubKey, "wrong-issuer", "aud")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

func TestParseAccessToken_WrongAudience(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	tokenStr, err := GenerateAccessToken("user-1", "sess-1", "iss", "correct-aud", privKey, 15*time.Minute)
	require.NoError(t, err)

	_, err = ParseAccessToken(tokenStr, pubKey, "iss", "wrong-aud")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

func TestParseAccessToken_MalformedToken(t *testing.T) {
	_, pubKey := generateTestRSAKeys(t)

	_, err := ParseAccessToken("not.a.valid.token", pubKey, "iss", "aud")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

func TestTokenService_GenerateAndValidate(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)
	svc := NewTokenService(privKey, pubKey, "test-issuer", "test-audience")

	pair, refreshToken, err := svc.GenerateTokenPair("user-abc", "session-xyz", "client-1")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, "Bearer", pair.TokenType)
	assert.Equal(t, int64(900), pair.ExpiresIn)
	assert.NotNil(t, refreshToken)

	// Validate access token
	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user-abc", claims.UserID)
	assert.Equal(t, "session-xyz", claims.SessionID)
}
