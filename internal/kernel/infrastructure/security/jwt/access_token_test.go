package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testIss = "test-issuer"
	testAud = "test-audience"
	testKid = "test-kid-1"
	testTTL = 15 * time.Minute
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
	}{
		{name: "happy_path", userID: "user-123", sessionID: "session-456"},
		{name: "empty_user_id_still_signs", userID: "", sessionID: "session-456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenStr, err := GenerateAccessToken(tt.userID, tt.sessionID, testIss, testAud, testKid, privKey, testTTL)
			require.NoError(t, err)
			require.NotEmpty(t, tokenStr)

			claims, err := ParseAccessToken(tokenStr, pubKey, testIss, testAud, 0)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, claims.Subject)
			assert.Equal(t, tt.sessionID, claims.SessionID)
			assert.Equal(t, TokenTypeAccess, claims.Type)
			assert.Equal(t, testIss, claims.Issuer)
			assert.Contains(t, []string(claims.Audience), testAud)
			assert.NotEmpty(t, claims.ID) // jti
		})
	}
}

func TestGenerateAccessToken_RejectsMissingInputs(t *testing.T) {
	privKey, _ := generateTestRSAKeys(t)

	_, err := GenerateAccessToken("u", "s", "", testAud, testKid, privKey, testTTL)
	assert.Error(t, err)

	_, err = GenerateAccessToken("u", "s", testIss, "", testKid, privKey, testTTL)
	assert.Error(t, err)

	_, err = GenerateAccessToken("u", "s", testIss, testAud, testKid, nil, testTTL)
	assert.Error(t, err)
}

func TestGenerateAccessToken_KidHeader(t *testing.T) {
	privKey, _ := generateTestRSAKeys(t)

	// With kid.
	tokenStr, err := GenerateAccessToken("u", "s", testIss, testAud, testKid, privKey, testTTL)
	require.NoError(t, err)
	parsed, _, err := jwtgo.NewParser().ParseUnverified(tokenStr, jwtgo.MapClaims{})
	require.NoError(t, err)
	assert.Equal(t, testKid, parsed.Header[HeaderKid])

	// Without kid — header must not be set.
	tokenStr, err = GenerateAccessToken("u", "s", testIss, testAud, "", privKey, testTTL)
	require.NoError(t, err)
	parsed, _, err = jwtgo.NewParser().ParseUnverified(tokenStr, jwtgo.MapClaims{})
	require.NoError(t, err)
	_, hasKid := parsed.Header[HeaderKid]
	assert.False(t, hasKid)
}

func TestParseAccessToken_Expired(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	// Emit a token that expired 1 hour ago (well beyond any leeway).
	tokenStr, err := GenerateAccessToken("user-1", "sess-1", testIss, testAud, testKid, privKey, -time.Hour)
	require.NoError(t, err)

	_, err = ParseAccessToken(tokenStr, pubKey, testIss, testAud, 0)
	assert.ErrorIs(t, err, ErrAccessTokenExpired)
}

func TestParseAccessToken_WithinLeeway(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	// Expired 10s ago, but leeway is 60s -> should be accepted.
	tokenStr, err := GenerateAccessToken("u", "s", testIss, testAud, testKid, privKey, -10*time.Second)
	require.NoError(t, err)

	_, err = ParseAccessToken(tokenStr, pubKey, testIss, testAud, 60*time.Second)
	assert.NoError(t, err)
}

func TestParseAccessToken_WrongKey(t *testing.T) {
	privKey1, _ := generateTestRSAKeys(t)
	_, pubKey2 := generateTestRSAKeys(t)

	tokenStr, err := GenerateAccessToken("u", "s", testIss, testAud, testKid, privKey1, testTTL)
	require.NoError(t, err)

	_, err = ParseAccessToken(tokenStr, pubKey2, testIss, testAud, 0)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

func TestParseAccessToken_WrongIssuer(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	tokenStr, err := GenerateAccessToken("u", "s", "correct-iss", testAud, testKid, privKey, testTTL)
	require.NoError(t, err)

	_, err = ParseAccessToken(tokenStr, pubKey, "wrong-iss", testAud, 0)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

func TestParseAccessToken_WrongAudience(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	tokenStr, err := GenerateAccessToken("u", "s", testIss, "correct-aud", testKid, privKey, testTTL)
	require.NoError(t, err)

	_, err = ParseAccessToken(tokenStr, pubKey, testIss, "wrong-aud", 0)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

func TestParseAccessToken_Malformed(t *testing.T) {
	_, pubKey := generateTestRSAKeys(t)

	_, err := ParseAccessToken("not.a.valid.token", pubKey, testIss, testAud, 0)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

// TestParseAccessToken_RejectsAlgDowngrade defends against the classic
// "HS256 using the RSA public key as HMAC secret" attack.
func TestParseAccessToken_RejectsAlgDowngrade(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	// Marshal the public key to PKIX bytes — these are the same bytes an
	// attacker could obtain from a JWKS endpoint.
	pubDER, err := x509.MarshalPKIXPublicKey(pubKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

	claims := AccessTokenClaims{
		SessionID: "s",
		Type:      TokenTypeAccess,
		RegisteredClaims: jwtgo.RegisteredClaims{
			Issuer:    testIss,
			Subject:   "u",
			Audience:  jwtgo.ClaimStrings{testAud},
			ExpiresAt: jwtgo.NewNumericDate(time.Now().UTC().Add(testTTL)),
			IssuedAt:  jwtgo.NewNumericDate(time.Now().UTC()),
		},
	}
	malicious := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	// Sign with HS256 using the PUBLIC KEY as the HMAC secret.
	evilToken, err := malicious.SignedString(pubPEM)
	require.NoError(t, err)

	// Parse must refuse — WithValidMethods enforces RS256.
	_, err = ParseAccessToken(evilToken, privKey.Public().(*rsa.PublicKey), testIss, testAud, 0)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

// TestParseAccessToken_RejectsNoneAlg defends against the "alg: none" attack.
func TestParseAccessToken_RejectsNoneAlg(t *testing.T) {
	_, pubKey := generateTestRSAKeys(t)

	claims := AccessTokenClaims{
		SessionID: "s",
		Type:      TokenTypeAccess,
		RegisteredClaims: jwtgo.RegisteredClaims{
			Issuer:    testIss,
			Subject:   "u",
			Audience:  jwtgo.ClaimStrings{testAud},
			ExpiresAt: jwtgo.NewNumericDate(time.Now().UTC().Add(testTTL)),
			IssuedAt:  jwtgo.NewNumericDate(time.Now().UTC()),
		},
	}
	noneTok := jwtgo.NewWithClaims(jwtgo.SigningMethodNone, claims)
	evil, err := noneTok.SignedString(jwtgo.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = ParseAccessToken(evil, pubKey, testIss, testAud, 0)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

// TestParseAccessToken_RejectsWrongType ensures a refresh-token-style
// "typ" claim cannot be smuggled through the access-token parser.
func TestParseAccessToken_RejectsWrongType(t *testing.T) {
	privKey, pubKey := generateTestRSAKeys(t)

	claims := AccessTokenClaims{
		SessionID: "s",
		Type:      TokenTypeRefresh, // wrong type
		RegisteredClaims: jwtgo.RegisteredClaims{
			Issuer:    testIss,
			Subject:   "u",
			Audience:  jwtgo.ClaimStrings{testAud},
			ExpiresAt: jwtgo.NewNumericDate(time.Now().UTC().Add(testTTL)),
			IssuedAt:  jwtgo.NewNumericDate(time.Now().UTC()),
		},
	}
	tok := jwtgo.NewWithClaims(jwtgo.SigningMethodRS256, claims)
	signed, err := tok.SignedString(privKey)
	require.NoError(t, err)

	_, err = ParseAccessToken(signed, pubKey, testIss, testAud, 0)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}
