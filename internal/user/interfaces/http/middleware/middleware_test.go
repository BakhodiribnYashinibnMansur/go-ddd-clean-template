package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/config"
	shared "gct/internal/shared/domain"
	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/security/jwt"
	"gct/internal/user/application/query"
	"gct/internal/user/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// nopLog implements logger.Log with no-ops for all methods.
// ---------------------------------------------------------------------------

type nopLog struct{}

func (nopLog) Debug(_ ...any)                                {}
func (nopLog) Debugf(_ string, _ ...any)                     {}
func (nopLog) Debugw(_ string, _ ...any)                     {}
func (nopLog) Info(_ ...any)                                 {}
func (nopLog) Infof(_ string, _ ...any)                      {}
func (nopLog) Infow(_ string, _ ...any)                      {}
func (nopLog) Warn(_ ...any)                                 {}
func (nopLog) Warnf(_ string, _ ...any)                      {}
func (nopLog) Warnw(_ string, _ ...any)                      {}
func (nopLog) Error(_ ...any)                                {}
func (nopLog) Errorf(_ string, _ ...any)                     {}
func (nopLog) Errorw(_ string, _ ...any)                     {}
func (nopLog) Fatal(_ ...any)                                {}
func (nopLog) Fatalf(_ string, _ ...any)                     {}
func (nopLog) Fatalw(_ string, _ ...any)                     {}
func (nopLog) Debugc(_ context.Context, _ string, _ ...any)  {}
func (nopLog) Infoc(_ context.Context, _ string, _ ...any)   {}
func (nopLog) Warnc(_ context.Context, _ string, _ ...any)   {}
func (nopLog) Errorc(_ context.Context, _ string, _ ...any)  {}
func (nopLog) Fatalc(_ context.Context, _ string, _ ...any)  {}

// ---------------------------------------------------------------------------
// fakeReadRepo implements domain.UserReadRepository with controllable returns.
// ---------------------------------------------------------------------------

type fakeReadRepo struct {
	session *shared.AuthSession
	sessErr error
	user    *shared.AuthUser
	userErr error
}

func (f *fakeReadRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.UserView, error) {
	return nil, nil
}

func (f *fakeReadRepo) List(_ context.Context, _ domain.UsersFilter) ([]*domain.UserView, int64, error) {
	return nil, 0, nil
}

func (f *fakeReadRepo) FindSessionByID(_ context.Context, _ uuid.UUID) (*shared.AuthSession, error) {
	return f.session, f.sessErr
}

func (f *fakeReadRepo) FindUserForAuth(_ context.Context, _ uuid.UUID) (*shared.AuthUser, error) {
	return f.user, f.userErr
}

// Ensure fakeReadRepo satisfies the interface at compile time.
var _ domain.UserReadRepository = (*fakeReadRepo)(nil)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func init() {
	gin.SetMode(gin.TestMode)
}

// generateRSAKeyPair returns a 2048-bit RSA key pair and the PEM-encoded
// public key string (suitable for config.Config.JWT.PublicKey).
func generateRSAKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey, string) {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1})
	return privKey, &privKey.PublicKey, string(pubPEM)
}

// newTestMiddleware constructs an AuthMiddleware wired with the given fake repo
// and a pre-parsed RSA public key (bypasses NewAuthMiddleware's config parsing).
func newTestMiddleware(t *testing.T, repo *fakeReadRepo, pubKey *rsa.PublicKey, issuer string) *AuthMiddleware {
	t.Helper()
	l := nopLog{}

	findSession := query.NewFindSessionHandler(repo, l)
	findUserForAuth := query.NewFindUserForAuthHandler(repo, l)

	return &AuthMiddleware{
		findSession:     findSession,
		findUserForAuth: findUserForAuth,
		cfg:             &config.Config{JWT: config.JWT{Issuer: issuer}},
		l:               l,
		pubKey:          pubKey,
	}
}

// newGinContext creates a fresh *gin.Context backed by an httptest.ResponseRecorder.
func newGinContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(method, path, nil)
	return ctx, w
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAuthApiKey_ReturnsForbidden(t *testing.T) {
	privKey, pubKey, _ := generateRSAKeyPair(t)
	_ = privKey // not needed for this test

	repo := &fakeReadRepo{}
	mw := newTestMiddleware(t, repo, pubKey, "test-issuer")

	// Case 1: no API key header -> 401 (key missing)
	ctx, w := newGinContext(http.MethodGet, "/")
	mw.AuthApiKey(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when API key missing, got %d", w.Code)
	}

	// Case 2: with API key header -> 403 (deprecated, always forbidden)
	ctx2, w2 := newGinContext(http.MethodGet, "/")
	ctx2.Request.Header.Set(consts.HeaderXAPIKey, "some-key-value")
	mw.AuthApiKey(ctx2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when API key present (deprecated), got %d", w2.Code)
	}
}

func TestParseAndValidateMetadata_ValidToken(t *testing.T) {
	const issuer = "test-issuer"
	privKey, pubKey, _ := generateRSAKeyPair(t)

	repo := &fakeReadRepo{}
	mw := newTestMiddleware(t, repo, pubKey, issuer)

	userID := uuid.New().String()
	sessionID := uuid.New().String()

	tokenStr, err := jwt.GenerateAccessToken(userID, sessionID, issuer, "", privKey, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	claims, err := mw.parseAndValidateMetadata(tokenStr)
	if err != nil {
		t.Fatalf("parseAndValidateMetadata returned unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %q, want %q", claims.UserID, userID)
	}
	if claims.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", claims.SessionID, sessionID)
	}
	if claims.Issuer != issuer {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, issuer)
	}
	if claims.Type != consts.TokenAccessType {
		t.Errorf("Type = %q, want %q", claims.Type, consts.TokenAccessType)
	}
}

func TestParseAndValidateMetadata_InvalidToken(t *testing.T) {
	_, pubKey, _ := generateRSAKeyPair(t)
	repo := &fakeReadRepo{}
	mw := newTestMiddleware(t, repo, pubKey, "test-issuer")

	_, err := mw.parseAndValidateMetadata("this-is-not-a-jwt")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestParseAndValidateMetadata_ExpiredToken(t *testing.T) {
	const issuer = "test-issuer"
	privKey, pubKey, _ := generateRSAKeyPair(t)

	repo := &fakeReadRepo{}
	mw := newTestMiddleware(t, repo, pubKey, issuer)

	// Generate a token that expired 1 hour ago.
	tokenStr, err := jwt.GenerateAccessToken(
		uuid.New().String(),
		uuid.New().String(),
		issuer,
		"",
		privKey,
		-1*time.Hour, // negative TTL -> already expired
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	_, err = mw.parseAndValidateMetadata(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestAuthClientAccess_NoToken(t *testing.T) {
	_, pubKey, _ := generateRSAKeyPair(t)
	repo := &fakeReadRepo{}
	mw := newTestMiddleware(t, repo, pubKey, "test-issuer")

	ctx, w := newGinContext(http.MethodGet, "/")
	mw.AuthClientAccess(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when no token present, got %d", w.Code)
	}
	if !ctx.IsAborted() {
		t.Fatal("expected context to be aborted")
	}
}

func TestAuthClientAccess_InvalidBearerFormat(t *testing.T) {
	_, pubKey, _ := generateRSAKeyPair(t)
	repo := &fakeReadRepo{}
	mw := newTestMiddleware(t, repo, pubKey, "test-issuer")

	ctx, w := newGinContext(http.MethodGet, "/")
	// "Bearer " with no actual token value -> ExtractBearerToken returns ""
	ctx.Request.Header.Set("Authorization", "Bearer ")
	mw.AuthClientAccess(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for malformed bearer, got %d", w.Code)
	}
	if !ctx.IsAborted() {
		t.Fatal("expected context to be aborted")
	}
}

func TestAuthAdmin_NoToken(t *testing.T) {
	_, pubKey, _ := generateRSAKeyPair(t)
	repo := &fakeReadRepo{}
	mw := newTestMiddleware(t, repo, pubKey, "test-issuer")

	ctx, w := newGinContext(http.MethodGet, "/admin")
	mw.AuthAdmin(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when no token present for admin, got %d", w.Code)
	}
	if !ctx.IsAborted() {
		t.Fatal("expected context to be aborted")
	}
}
