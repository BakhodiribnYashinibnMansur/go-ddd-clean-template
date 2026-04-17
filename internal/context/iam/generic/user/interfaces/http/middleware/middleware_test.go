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
	"gct/internal/context/iam/generic/user/application/query"
	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	"gct/internal/kernel/consts"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/security/audit"
	"gct/internal/kernel/infrastructure/security/fingerprint"
	"gct/internal/kernel/infrastructure/security/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// nopLog implements logger.Log with no-ops for all methods.
// ---------------------------------------------------------------------------

type nopLog struct{}

func (nopLog) Debug(_ ...any)                               {}
func (nopLog) Debugf(_ string, _ ...any)                    {}
func (nopLog) Debugw(_ string, _ ...any)                    {}
func (nopLog) Info(_ ...any)                                {}
func (nopLog) Infof(_ string, _ ...any)                     {}
func (nopLog) Infow(_ string, _ ...any)                     {}
func (nopLog) Warn(_ ...any)                                {}
func (nopLog) Warnf(_ string, _ ...any)                     {}
func (nopLog) Warnw(_ string, _ ...any)                     {}
func (nopLog) Error(_ ...any)                               {}
func (nopLog) Errorf(_ string, _ ...any)                    {}
func (nopLog) Errorw(_ string, _ ...any)                    {}
func (nopLog) Fatal(_ ...any)                               {}
func (nopLog) Fatalf(_ string, _ ...any)                    {}
func (nopLog) Fatalw(_ string, _ ...any)                    {}
func (nopLog) Debugc(_ context.Context, _ string, _ ...any) {}
func (nopLog) Infoc(_ context.Context, _ string, _ ...any)  {}
func (nopLog) Warnc(_ context.Context, _ string, _ ...any)  {}
func (nopLog) Errorc(_ context.Context, _ string, _ ...any) {}
func (nopLog) Fatalc(_ context.Context, _ string, _ ...any) {}

// ---------------------------------------------------------------------------
// fakeReadRepo implements userrepo.UserReadRepository with controllable returns.
// ---------------------------------------------------------------------------

type fakeReadRepo struct {
	session *shared.AuthSession
	sessErr error
	user    *shared.AuthUser
	userErr error
}

func (f *fakeReadRepo) FindByID(_ context.Context, _ userentity.UserID) (*userentity.UserView, error) {
	return nil, nil
}

func (f *fakeReadRepo) List(_ context.Context, _ userentity.UsersFilter) ([]*userentity.UserView, int64, error) {
	return nil, 0, nil
}

func (f *fakeReadRepo) FindSessionByID(_ context.Context, _ userentity.SessionID) (*shared.AuthSession, error) {
	return f.session, f.sessErr
}

func (f *fakeReadRepo) FindUserForAuth(_ context.Context, _ userentity.UserID) (*shared.AuthUser, error) {
	return f.user, f.userErr
}

func (f *fakeReadRepo) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
}

// Ensure fakeReadRepo satisfies the interface at compile time.
var _ userrepo.UserReadRepository = (*fakeReadRepo)(nil)

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

// fakeResolver is a test double for IntegrationResolver returning a fixed
// ResolvedForVerify.
type fakeResolver struct {
	resolved *ResolvedForVerify
	err      error
}

func (f *fakeResolver) Resolve(_ context.Context, _ string) (*ResolvedForVerify, error) {
	return f.resolved, f.err
}

// fakeSessionRevoker records calls to RevokeSessionsByIntegration.
type fakeSessionRevoker struct {
	called          bool
	revokedUserID   uuid.UUID
	revokedIntName  string
	revokedCount    int
	err             error
}

func (f *fakeSessionRevoker) RevokeSessionsByIntegration(_ context.Context, userID uuid.UUID, integrationName string) (int, error) {
	f.called = true
	f.revokedUserID = userID
	f.revokedIntName = integrationName
	return f.revokedCount, f.err
}

// newTestMiddleware constructs an AuthMiddleware wired with the given fake repo
// and a pre-parsed RSA public key (bypasses NewAuthMiddleware's config parsing).
func newTestMiddleware(t *testing.T, repo *fakeReadRepo, pubKey *rsa.PublicKey, issuer string) *AuthMiddleware {
	t.Helper()
	return newTestMiddlewareWithRevoker(t, repo, pubKey, issuer, nil)
}

// newTestMiddlewareWithRevoker is like newTestMiddleware but accepts an optional
// SessionRevoker for reuse detection tests.
func newTestMiddlewareWithRevoker(t *testing.T, repo *fakeReadRepo, pubKey *rsa.PublicKey, issuer string, revoker SessionRevoker) *AuthMiddleware {
	t.Helper()
	l := nopLog{}

	findSession := query.NewFindSessionHandler(repo, l)
	findUserForAuth := query.NewFindUserForAuthHandler(repo, l)

	// Fixed 32-byte pepper for tests. The hasher copies it internally.
	hasher, err := jwt.NewRefreshHasher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("jwt.NewRefreshHasher: %v", err)
	}

	resolver := &fakeResolver{resolved: &ResolvedForVerify{
		Name:      "test-aud",
		PublicKey: pubKey,
		KeyID:     "test-kid",
	}}

	return &AuthMiddleware{
		findSession:     findSession,
		findUserForAuth: findUserForAuth,
		cfg:             &config.Config{JWT: config.JWT{Issuer: issuer, Leeway: 30 * time.Second}},
		l:               l,
		resolver:        resolver,
		refreshHasher:   hasher,
		sessionRevoker:  revoker,
		auditLogger:     audit.NoopLogger{},
		issuer:          issuer,
		leeway:          30 * time.Second,
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

	userID := userentity.NewUserID().String()
	sessionID := userentity.NewSessionID().String()

	tokenStr, err := jwt.GenerateAccessToken(userID, sessionID, issuer, "test-aud", "", privKey, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	claims, err := mw.parseAndValidateMetadata(tokenStr, &ResolvedForVerify{Name: "test-aud", PublicKey: pubKey})
	if err != nil {
		t.Fatalf("parseAndValidateMetadata returned unexpected error: %v", err)
	}
	if claims.Subject != userID {
		t.Errorf("Subject = %q, want %q", claims.Subject, userID)
	}
	if claims.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", claims.SessionID, sessionID)
	}
	if claims.Issuer != issuer {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, issuer)
	}
	if claims.Type != jwt.TokenTypeAccess {
		t.Errorf("Type = %q, want %q", claims.Type, jwt.TokenTypeAccess)
	}
}

func TestParseAndValidateMetadata_InvalidToken(t *testing.T) {
	_, pubKey, _ := generateRSAKeyPair(t)
	repo := &fakeReadRepo{}
	mw := newTestMiddleware(t, repo, pubKey, "test-issuer")

	_, err := mw.parseAndValidateMetadata("this-is-not-a-jwt", &ResolvedForVerify{Name: "test-aud", PublicKey: pubKey})
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
		"test-aud",
		"",
		privKey,
		-1*time.Hour, // negative TTL -> already expired
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	_, err = mw.parseAndValidateMetadata(tokenStr, &ResolvedForVerify{Name: "test-aud", PublicKey: pubKey})
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

// TestValidateAccessToken_CrossIntegrationRejected asserts that a session
// bound to one integration cannot be used by a token minted for a different
// integration (audience mismatch at the session level, post-JWT-verify).
func TestValidateAccessToken_CrossIntegrationRejected(t *testing.T) {
	const issuer = "test-issuer"
	privKey, pubKey, _ := generateRSAKeyPair(t)

	sessUUID := uuid.New()
	userUUID := uuid.New()
	// Session was bound to gct-client.
	repo := &fakeReadRepo{session: &shared.AuthSession{
		ID: sessUUID, UserID: userUUID,
		IntegrationName: "gct-client",
		ExpiresAt:       time.Now().Add(1 * time.Hour),
	}}
	mw := newTestMiddleware(t, repo, pubKey, issuer)
	// Swap the resolver to return a *different* audience, simulating a
	// token presented by a crafted X-API-Key for gct-admin.
	mw.resolver = &fakeResolver{resolved: &ResolvedForVerify{
		Name: "gct-admin", PublicKey: pubKey, KeyID: "admin-kid",
	}}

	// Generate a valid-looking token for audience gct-admin.
	tokenStr, err := jwt.GenerateAccessToken(userUUID.String(), sessUUID.String(), issuer, "gct-admin", "admin-kid", privKey, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	ctx, w := newGinContext(http.MethodGet, "/")
	ctx.Request.Header.Set("Authorization", "Bearer "+tokenStr)
	ctx.Request.Header.Set(consts.HeaderXAPIKey, "some-key")
	mw.AuthClientAccess(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for cross-integration token, got %d", w.Code)
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

// ---------------------------------------------------------------------------
// Refresh-token rotation & reuse detection
// ---------------------------------------------------------------------------

// buildRefreshToken creates a refresh token string and its hash for testing.
func buildRefreshToken(t *testing.T, hasher *jwt.RefreshHasher, sessionID string) (tokenStr, hash string) {
	t.Helper()
	rt, err := jwt.GenerateRefreshToken(hasher, uuid.New().String(), sessionID, "device-1", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	return rt.String(), rt.Hashed
}

func TestAuthClientRefresh_CurrentHashMatches(t *testing.T) {
	t.Parallel()
	_, pubKey, _ := generateRSAKeyPair(t)

	hasher, err := jwt.NewRefreshHasher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewRefreshHasher: %v", err)
	}

	sessUUID := uuid.New()
	userUUID := uuid.New()
	tokenStr, hash := buildRefreshToken(t, hasher, sessUUID.String())

	repo := &fakeReadRepo{session: &shared.AuthSession{
		ID: sessUUID, UserID: userUUID,
		RefreshTokenHash: hash,
		IntegrationName:  "test-aud",
		ExpiresAt:        time.Now().Add(1 * time.Hour),
	}}
	revoker := &fakeSessionRevoker{revokedCount: 3}
	mw := newTestMiddlewareWithRevoker(t, repo, pubKey, "test-issuer", revoker)

	ctx, w := newGinContext(http.MethodPost, "/auth/refresh")
	ctx.Request.Header.Set(consts.HeaderXAPIKey, "some-key")
	ctx.Request.Header.Set("Authorization", "Bearer "+tokenStr)
	// AuthClientRefresh calls ctx.Next(); we need a handler registered.
	// In the test harness ctx.Next() is a no-op, which is fine.
	mw.AuthClientRefresh(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid current hash, got %d", w.Code)
	}
	if ctx.IsAborted() {
		t.Fatal("context should not be aborted for valid refresh")
	}
	if revoker.called {
		t.Fatal("revoker should not be called when current hash matches")
	}
	// Verify context vars were set.
	if ctx.GetString(consts.CtxUserID) != userUUID.String() {
		t.Fatal("CtxUserID not set correctly")
	}
}

func TestAuthClientRefresh_PreviousHashMatches_ReuseDetected(t *testing.T) {
	t.Parallel()
	_, pubKey, _ := generateRSAKeyPair(t)

	hasher, err := jwt.NewRefreshHasher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewRefreshHasher: %v", err)
	}

	sessUUID := uuid.New()
	userUUID := uuid.New()

	// This token was already rotated: its hash is now in previous_refresh_hash,
	// and current hash is something new.
	tokenStr, oldHash := buildRefreshToken(t, hasher, sessUUID.String())

	repo := &fakeReadRepo{session: &shared.AuthSession{
		ID: sessUUID, UserID: userUUID,
		RefreshTokenHash:    "new-current-hash-after-rotation",
		PreviousRefreshHash: oldHash,
		IntegrationName:     "test-aud",
		ExpiresAt:           time.Now().Add(1 * time.Hour),
	}}
	revoker := &fakeSessionRevoker{revokedCount: 3}
	mw := newTestMiddlewareWithRevoker(t, repo, pubKey, "test-issuer", revoker)

	ctx, w := newGinContext(http.MethodPost, "/auth/refresh")
	ctx.Request.Header.Set(consts.HeaderXAPIKey, "some-key")
	ctx.Request.Header.Set("Authorization", "Bearer "+tokenStr)
	mw.AuthClientRefresh(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for reuse detection, got %d", w.Code)
	}
	if !ctx.IsAborted() {
		t.Fatal("context should be aborted on reuse detection")
	}
	if !revoker.called {
		t.Fatal("revoker should have been called on reuse detection")
	}
	if revoker.revokedUserID != userUUID {
		t.Fatalf("revoker received wrong user ID: got %s, want %s", revoker.revokedUserID, userUUID)
	}
	if revoker.revokedIntName != "test-aud" {
		t.Fatalf("revoker received wrong integration: got %s, want test-aud", revoker.revokedIntName)
	}
}

func TestAuthClientRefresh_NoHashMatches(t *testing.T) {
	t.Parallel()
	_, pubKey, _ := generateRSAKeyPair(t)

	hasher, err := jwt.NewRefreshHasher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewRefreshHasher: %v", err)
	}

	sessUUID := uuid.New()
	userUUID := uuid.New()
	tokenStr, _ := buildRefreshToken(t, hasher, sessUUID.String())

	repo := &fakeReadRepo{session: &shared.AuthSession{
		ID: sessUUID, UserID: userUUID,
		RefreshTokenHash:    "completely-different-hash",
		PreviousRefreshHash: "also-different-previous",
		IntegrationName:     "test-aud",
		ExpiresAt:           time.Now().Add(1 * time.Hour),
	}}
	revoker := &fakeSessionRevoker{}
	mw := newTestMiddlewareWithRevoker(t, repo, pubKey, "test-issuer", revoker)

	ctx, w := newGinContext(http.MethodPost, "/auth/refresh")
	ctx.Request.Header.Set(consts.HeaderXAPIKey, "some-key")
	ctx.Request.Header.Set("Authorization", "Bearer "+tokenStr)
	mw.AuthClientRefresh(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", w.Code)
	}
	if !ctx.IsAborted() {
		t.Fatal("context should be aborted for invalid token")
	}
	if revoker.called {
		t.Fatal("revoker should not be called when neither hash matches")
	}
}

// ---------------------------------------------------------------------------
// Device fingerprint binding
// ---------------------------------------------------------------------------

func TestAuthClientRefresh_FingerprintMismatch_Returns401(t *testing.T) {
	t.Parallel()
	_, pubKey, _ := generateRSAKeyPair(t)

	hasher, err := jwt.NewRefreshHasher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewRefreshHasher: %v", err)
	}

	sessUUID := uuid.New()
	userUUID := uuid.New()
	tokenStr, hash := buildRefreshToken(t, hasher, sessUUID.String())

	// Session was created with a specific device fingerprint.
	originalFP := fingerprint.Compute("OriginalUA/1.0", "en-US", "Chromium")
	repo := &fakeReadRepo{session: &shared.AuthSession{
		ID: sessUUID, UserID: userUUID,
		RefreshTokenHash:  hash,
		IntegrationName:   "test-aud",
		ExpiresAt:         time.Now().Add(1 * time.Hour),
		DeviceFingerprint: originalFP,
	}}
	revoker := &fakeSessionRevoker{}
	mw := newTestMiddlewareWithRevoker(t, repo, pubKey, "test-issuer", revoker)

	ctx, w := newGinContext(http.MethodPost, "/auth/refresh")
	ctx.Request.Header.Set(consts.HeaderXAPIKey, "some-key")
	ctx.Request.Header.Set("Authorization", "Bearer "+tokenStr)
	// Present a DIFFERENT user agent -> different fingerprint.
	ctx.Request.Header.Set("User-Agent", "DifferentUA/2.0")
	ctx.Request.Header.Set("Accept-Language", "fr-FR")
	mw.AuthClientRefresh(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for fingerprint mismatch, got %d", w.Code)
	}
	if !ctx.IsAborted() {
		t.Fatal("context should be aborted on fingerprint mismatch")
	}
}

func TestAuthClientRefresh_FingerprintMatch_Passes(t *testing.T) {
	t.Parallel()
	_, pubKey, _ := generateRSAKeyPair(t)

	hasher, err := jwt.NewRefreshHasher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewRefreshHasher: %v", err)
	}

	sessUUID := uuid.New()
	userUUID := uuid.New()
	tokenStr, hash := buildRefreshToken(t, hasher, sessUUID.String())

	// Session fingerprint matches the headers we will send.
	fp := fingerprint.Compute("TestUA/1.0", "en-US", "")
	repo := &fakeReadRepo{session: &shared.AuthSession{
		ID: sessUUID, UserID: userUUID,
		RefreshTokenHash:  hash,
		IntegrationName:   "test-aud",
		ExpiresAt:         time.Now().Add(1 * time.Hour),
		DeviceFingerprint: fp,
	}}
	revoker := &fakeSessionRevoker{}
	mw := newTestMiddlewareWithRevoker(t, repo, pubKey, "test-issuer", revoker)

	ctx, w := newGinContext(http.MethodPost, "/auth/refresh")
	ctx.Request.Header.Set(consts.HeaderXAPIKey, "some-key")
	ctx.Request.Header.Set("Authorization", "Bearer "+tokenStr)
	ctx.Request.Header.Set("User-Agent", "TestUA/1.0")
	ctx.Request.Header.Set("Accept-Language", "en-US")
	mw.AuthClientRefresh(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for matching fingerprint, got %d", w.Code)
	}
	if ctx.IsAborted() {
		t.Fatal("context should not be aborted when fingerprint matches")
	}
}

func TestAuthClientRefresh_EmptyFingerprint_Skips(t *testing.T) {
	t.Parallel()
	_, pubKey, _ := generateRSAKeyPair(t)

	hasher, err := jwt.NewRefreshHasher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewRefreshHasher: %v", err)
	}

	sessUUID := uuid.New()
	userUUID := uuid.New()
	tokenStr, hash := buildRefreshToken(t, hasher, sessUUID.String())

	// No fingerprint stored on session — check should be skipped.
	repo := &fakeReadRepo{session: &shared.AuthSession{
		ID: sessUUID, UserID: userUUID,
		RefreshTokenHash: hash,
		IntegrationName:  "test-aud",
		ExpiresAt:        time.Now().Add(1 * time.Hour),
	}}
	revoker := &fakeSessionRevoker{}
	mw := newTestMiddlewareWithRevoker(t, repo, pubKey, "test-issuer", revoker)

	ctx, w := newGinContext(http.MethodPost, "/auth/refresh")
	ctx.Request.Header.Set(consts.HeaderXAPIKey, "some-key")
	ctx.Request.Header.Set("Authorization", "Bearer "+tokenStr)
	ctx.Request.Header.Set("User-Agent", "AnyAgent/99")
	mw.AuthClientRefresh(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when session has no fingerprint, got %d", w.Code)
	}
	if ctx.IsAborted() {
		t.Fatal("context should not be aborted when no fingerprint is stored")
	}
}
