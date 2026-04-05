package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"gct/internal/platform/domain/consts"

	gojwt "github.com/golang-jwt/jwt/v4"
)

func testRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return key
}

func TestGenerateToken_BasicClaims(t *testing.T) {
	key := testRSAKey(t)

	tokenStr, err := GenerateToken(TokenParams{
		Issuer:     "test-issuer",
		Subject:    "user-123",
		SessionID:  "session-456",
		Audience:   "test-audience",
		Type:       "access",
		TTL:        15 * time.Minute,
		PrivateKey: key,
	})
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}

	// Parse and verify claims
	token, err := gojwt.Parse(tokenStr, func(token *gojwt.Token) (any, error) {
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !token.Valid {
		t.Fatal("expected valid token")
	}

	claims, ok := token.Claims.(gojwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}
	if claims[consts.ClaimIssuer] != "test-issuer" {
		t.Errorf("expected issuer 'test-issuer', got %v", claims[consts.ClaimIssuer])
	}
	if claims[consts.ClaimSubject] != "user-123" {
		t.Errorf("expected subject 'user-123', got %v", claims[consts.ClaimSubject])
	}
	if claims[consts.ClaimSessionID] != "session-456" {
		t.Errorf("expected session ID 'session-456', got %v", claims[consts.ClaimSessionID])
	}
	if claims[consts.ClaimAudience] != "test-audience" {
		t.Errorf("expected audience 'test-audience', got %v", claims[consts.ClaimAudience])
	}
	if claims[consts.ClaimType] != "access" {
		t.Errorf("expected type 'access', got %v", claims[consts.ClaimType])
	}
}

func TestGenerateToken_OptionalClaims(t *testing.T) {
	key := testRSAKey(t)

	tokenStr, err := GenerateToken(TokenParams{
		Issuer:          "test-issuer",
		Subject:         "user-123",
		SessionID:       "session-456",
		Audience:        "test-audience",
		CompanyID:       "company-789",
		Scope:           []string{"read", "write"},
		AuthorizedParty: "my-app",
		Type:            "access",
		TTL:             15 * time.Minute,
		PrivateKey:      key,
	})
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	token, err := gojwt.Parse(tokenStr, func(token *gojwt.Token) (any, error) {
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims := token.Claims.(gojwt.MapClaims)
	if claims[consts.ClaimCompanyID] != "company-789" {
		t.Errorf("expected company ID 'company-789', got %v", claims[consts.ClaimCompanyID])
	}
	if claims[consts.ClaimAuthorizedParty] != "my-app" {
		t.Errorf("expected azp 'my-app', got %v", claims[consts.ClaimAuthorizedParty])
	}

	scopeRaw, ok := claims[consts.ClaimScope]
	if !ok {
		t.Fatal("expected scope claim to be present")
	}
	scopeSlice, ok := scopeRaw.([]any)
	if !ok {
		t.Fatalf("expected scope to be []any, got %T", scopeRaw)
	}
	if len(scopeSlice) != 2 {
		t.Fatalf("expected 2 scopes, got %d", len(scopeSlice))
	}
}

func TestGenerateToken_OmitsEmptyOptionalClaims(t *testing.T) {
	key := testRSAKey(t)

	tokenStr, err := GenerateToken(TokenParams{
		Issuer:     "test-issuer",
		Subject:    "user-123",
		SessionID:  "session-456",
		Audience:   "test-audience",
		Type:       "access",
		TTL:        15 * time.Minute,
		PrivateKey: key,
	})
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	token, err := gojwt.Parse(tokenStr, func(token *gojwt.Token) (any, error) {
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims := token.Claims.(gojwt.MapClaims)
	if _, ok := claims[consts.ClaimCompanyID]; ok {
		t.Error("expected company ID claim to be absent when empty")
	}
	if _, ok := claims[consts.ClaimScope]; ok {
		t.Error("expected scope claim to be absent when nil")
	}
	if _, ok := claims[consts.ClaimAuthorizedParty]; ok {
		t.Error("expected azp claim to be absent when empty")
	}
}

func TestGenerateToken_ExpirationIsSet(t *testing.T) {
	key := testRSAKey(t)
	ttl := 30 * time.Minute

	tokenStr, err := GenerateToken(TokenParams{
		Issuer:     "test-issuer",
		Subject:    "user-123",
		SessionID:  "session-456",
		Audience:   "test-audience",
		Type:       "access",
		TTL:        ttl,
		PrivateKey: key,
	})
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	token, err := gojwt.Parse(tokenStr, func(token *gojwt.Token) (any, error) {
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims := token.Claims.(gojwt.MapClaims)
	exp, ok := claims[consts.ClaimExpiresAt].(float64)
	if !ok {
		t.Fatal("expected exp claim to be a float64")
	}
	iat, ok := claims[consts.ClaimIssuedAt].(float64)
	if !ok {
		t.Fatal("expected iat claim to be a float64")
	}

	diff := exp - iat
	expectedDiff := ttl.Seconds()
	if diff < expectedDiff-1 || diff > expectedDiff+1 {
		t.Errorf("expected exp-iat to be ~%.0f, got %.0f", expectedDiff, diff)
	}
}

func TestGenerateToken_SigningMethodRS256(t *testing.T) {
	key := testRSAKey(t)

	tokenStr, err := GenerateToken(TokenParams{
		Issuer:     "test-issuer",
		Subject:    "user-123",
		SessionID:  "session-456",
		Audience:   "test-audience",
		Type:       "access",
		TTL:        15 * time.Minute,
		PrivateKey: key,
	})
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	token, err := gojwt.Parse(tokenStr, func(token *gojwt.Token) (any, error) {
		if _, ok := token.Method.(*gojwt.SigningMethodRSA); !ok {
			t.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !token.Valid {
		t.Error("expected valid token")
	}
}
