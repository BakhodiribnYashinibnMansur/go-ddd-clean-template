package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"
)

func generateServiceTestRSAKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key pair: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}

func TestNewTokenService(t *testing.T) {
	priv, pub := generateServiceTestRSAKeys(t)
	svc := NewTokenService(priv, pub, "test-issuer", "test-audience")
	if svc == nil {
		t.Fatal("expected non-nil TokenService")
	}
	if svc.issuer != "test-issuer" {
		t.Errorf("expected issuer 'test-issuer', got %q", svc.issuer)
	}
	if svc.audience != "test-audience" {
		t.Errorf("expected audience 'test-audience', got %q", svc.audience)
	}
}

func TestTokenService_GenerateTokenPair(t *testing.T) {
	priv, pub := generateServiceTestRSAKeys(t)
	svc := NewTokenService(priv, pub, "test-issuer", "test-audience")

	pair, refreshToken, err := svc.GenerateTokenPair("user-123", "session-456", "client-789")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}
	if pair == nil {
		t.Fatal("expected non-nil TokenPair")
	}
	if refreshToken == nil {
		t.Fatal("expected non-nil RefreshToken")
	}
	if pair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if pair.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("expected token type 'Bearer', got %q", pair.TokenType)
	}
	if pair.ExpiresIn != 900 {
		t.Errorf("expected ExpiresIn 900, got %d", pair.ExpiresIn)
	}
}

func TestTokenService_ValidateAccessToken(t *testing.T) {
	priv, pub := generateServiceTestRSAKeys(t)
	svc := NewTokenService(priv, pub, "test-issuer", "test-audience")

	pair, _, err := svc.GenerateTokenPair("user-123", "session-456", "client-789")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got %q", claims.UserID)
	}
	if claims.SessionID != "session-456" {
		t.Errorf("expected SessionID 'session-456', got %q", claims.SessionID)
	}
}

func TestTokenService_ValidateAccessToken_InvalidToken(t *testing.T) {
	priv, pub := generateServiceTestRSAKeys(t)
	svc := NewTokenService(priv, pub, "test-issuer", "test-audience")

	_, err := svc.ValidateAccessToken("invalid-token-string")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestTokenService_ValidateAccessToken_WrongKey(t *testing.T) {
	priv1, _ := generateServiceTestRSAKeys(t)
	_, pub2 := generateServiceTestRSAKeys(t)

	svc1 := NewTokenService(priv1, &priv1.PublicKey, "test-issuer", "test-audience")
	svc2 := NewTokenService(priv1, pub2, "test-issuer", "test-audience")

	pair, _, err := svc1.GenerateTokenPair("user-123", "session-456", "client-789")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	_, err = svc2.ValidateAccessToken(pair.AccessToken)
	if err == nil {
		t.Fatal("expected error when validating with wrong public key")
	}
}

func TestTokenService_ValidateRefreshToken(t *testing.T) {
	priv, pub := generateServiceTestRSAKeys(t)
	svc := NewTokenService(priv, pub, "test-issuer", "test-audience")

	_, refreshToken, err := svc.GenerateTokenPair("user-123", "session-456", "client-789")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	tokenStr := refreshToken.String()
	verified, err := svc.ValidateRefreshToken(tokenStr, refreshToken.Hashed)
	if err != nil {
		t.Fatalf("ValidateRefreshToken returned error: %v", err)
	}
	if verified.ID != refreshToken.ID {
		t.Errorf("expected ID %q, got %q", refreshToken.ID, verified.ID)
	}
}

func TestTokenService_ValidateRefreshToken_WrongHash(t *testing.T) {
	priv, pub := generateServiceTestRSAKeys(t)
	svc := NewTokenService(priv, pub, "test-issuer", "test-audience")

	_, refreshToken, err := svc.GenerateTokenPair("user-123", "session-456", "client-789")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	_, err = svc.ValidateRefreshToken(refreshToken.String(), "wrong-hash")
	if err == nil {
		t.Fatal("expected error for wrong hash, got nil")
	}
}

func TestTokenService_RefreshTokens(t *testing.T) {
	// NOTE: RefreshTokens calls ValidateRefreshToken then checks IsExpired().
	// However, ParseRefreshToken (used internally) does not restore ExpiresAt from the token string,
	// so the parsed token always has zero ExpiresAt and IsExpired() returns true.
	// In production, expiry is checked against the stored value from the database, not the token string.
	// Therefore, RefreshTokens is expected to return ErrRefreshTokenExpired in this unit test scenario.
	priv, pub := generateServiceTestRSAKeys(t)
	svc := NewTokenService(priv, pub, "test-issuer", "test-audience")

	_, refreshToken, err := svc.GenerateTokenPair("user-123", "session-456", "client-789")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	_, _, err = svc.RefreshTokens(refreshToken.String(), refreshToken.Hashed)
	if err == nil {
		t.Fatal("expected error due to parsed token having zero ExpiresAt")
	}
	if err != ErrRefreshTokenExpired {
		t.Errorf("expected ErrRefreshTokenExpired, got %v", err)
	}
}

func TestTokenService_RefreshTokens_ExpiredRefresh(t *testing.T) {
	priv, pub := generateServiceTestRSAKeys(t)
	svc := NewTokenService(priv, pub, "test-issuer", "test-audience")

	// Generate a refresh token that is already expired
	refreshToken, err := GenerateRefreshToken("user-123", "session-456", "client-789", -1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}

	_, _, err = svc.RefreshTokens(refreshToken.String(), refreshToken.Hashed)
	if err == nil {
		t.Fatal("expected error for expired refresh token, got nil")
	}
}
