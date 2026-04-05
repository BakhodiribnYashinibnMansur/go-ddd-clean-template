package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"gct/config"
	"gct/internal/context/admin/supporting/integration/application/query"
	"gct/internal/context/admin/supporting/integration/domain"
	"gct/internal/kernel/consts"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ---------------------------------------------------------------------------
// Mock IntegrationReadRepository
// ---------------------------------------------------------------------------

type mockReadRepo struct {
	apiKeyView *domain.IntegrationAPIKeyView
	findErr    error
}

func (r *mockReadRepo) FindByID(_ context.Context, _ domain.IntegrationID) (*domain.IntegrationView, error) {
	return nil, errors.New("not implemented")
}

func (r *mockReadRepo) List(_ context.Context, _ domain.IntegrationFilter) ([]*domain.IntegrationView, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (r *mockReadRepo) FindByAPIKey(_ context.Context, _ string) (*domain.IntegrationAPIKeyView, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.apiKeyView, nil
}

// ---------------------------------------------------------------------------
// Mock Logger
// ---------------------------------------------------------------------------

type mockLog struct{}

func (m *mockLog) Debug(_ ...any)                               {}
func (m *mockLog) Debugf(_ string, _ ...any)                    {}
func (m *mockLog) Debugw(_ string, _ ...any)                    {}
func (m *mockLog) Info(_ ...any)                                {}
func (m *mockLog) Infof(_ string, _ ...any)                     {}
func (m *mockLog) Infow(_ string, _ ...any)                     {}
func (m *mockLog) Warn(_ ...any)                                {}
func (m *mockLog) Warnf(_ string, _ ...any)                     {}
func (m *mockLog) Warnw(_ string, _ ...any)                     {}
func (m *mockLog) Error(_ ...any)                               {}
func (m *mockLog) Errorf(_ string, _ ...any)                    {}
func (m *mockLog) Errorw(_ string, _ ...any)                    {}
func (m *mockLog) Fatal(_ ...any)                               {}
func (m *mockLog) Fatalf(_ string, _ ...any)                    {}
func (m *mockLog) Fatalw(_ string, _ ...any)                    {}
func (m *mockLog) Debugc(_ context.Context, _ string, _ ...any) {}
func (m *mockLog) Infoc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLog) Warnc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLog) Errorc(_ context.Context, _ string, _ ...any) {}
func (m *mockLog) Fatalc(_ context.Context, _ string, _ ...any) {}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeSign(timeUnix, apiKey, requestID string) string {
	data := timeUnix + apiKey + requestID
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func newMiddleware(repo *mockReadRepo, expireTime int64) *SignatureMiddleware {
	handler := query.NewValidateAPIKeyHandler(repo, &mockLog{})
	cfg := &config.Config{}
	cfg.APIKeys.SignExpireTime = expireTime
	return NewSignatureMiddleware(handler, cfg)
}

func performRequest(mw gin.HandlerFunc, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	router := gin.New()
	router.Use(mw)
	router.Any("/*path", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// Tests: Skipped paths
// ---------------------------------------------------------------------------

func TestSignature_SkipAdminPath(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	w := performRequest(mw.Validate(), "GET", "/admin/dashboard", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /admin path, got %d", w.Code)
	}
}

func TestSignature_SkipStaticPath(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	w := performRequest(mw.Validate(), "GET", "/static/style.css", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /static path, got %d", w.Code)
	}
}

func TestSignature_SkipDocsPath(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	w := performRequest(mw.Validate(), "GET", "/docs/swagger", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /docs path, got %d", w.Code)
	}
}

func TestSignature_SkipHealthPath(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	w := performRequest(mw.Validate(), "GET", "/health", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /health path, got %d", w.Code)
	}
}

func TestSignature_SkipRootPath(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	w := performRequest(mw.Validate(), "GET", "/", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for / path, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: Missing headers
// ---------------------------------------------------------------------------

func TestSignature_MissingTimeUnix(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	w := performRequest(mw.Validate(), "GET", "/api/v1/test", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing X-Time-Unix, got %d", w.Code)
	}
}

func TestSignature_MissingRequestID(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	headers := map[string]string{
		consts.HeaderXTimeUnix: strconv.FormatInt(time.Now().Unix(), 10),
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing X-Request-ID, got %d", w.Code)
	}
}

func TestSignature_MissingAPIKey(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	headers := map[string]string{
		consts.HeaderXTimeUnix:  strconv.FormatInt(time.Now().Unix(), 10),
		consts.HeaderXRequestID: "req-123",
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing API key, got %d", w.Code)
	}
}

func TestSignature_MissingSign(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	headers := map[string]string{
		consts.HeaderXTimeUnix:  strconv.FormatInt(time.Now().Unix(), 10),
		consts.HeaderXRequestID: "req-123",
		consts.HeaderXAPIKey:    "my-api-key",
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing X-Sign, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: Timestamp validation
// ---------------------------------------------------------------------------

func TestSignature_InvalidTimeFormat(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	headers := map[string]string{
		consts.HeaderXTimeUnix:  "not-a-number",
		consts.HeaderXRequestID: "req-123",
		consts.HeaderXAPIKey:    "my-api-key",
		consts.HeaderXSign:      "somesign",
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid time format, got %d", w.Code)
	}
}

func TestSignature_ExpiredTimestamp(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	expiredTime := strconv.FormatInt(time.Now().Unix()-100, 10)
	headers := map[string]string{
		consts.HeaderXTimeUnix:  expiredTime,
		consts.HeaderXRequestID: "req-123",
		consts.HeaderXAPIKey:    "my-api-key",
		consts.HeaderXSign:      "somesign",
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	if w.Code != 499 {
		t.Errorf("expected 499 for expired timestamp, got %d", w.Code)
	}
}

func TestSignature_FutureTimestamp(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{}
	mw := newMiddleware(repo, 10)

	futureTime := strconv.FormatInt(time.Now().Unix()+100, 10)
	headers := map[string]string{
		consts.HeaderXTimeUnix:  futureTime,
		consts.HeaderXRequestID: "req-123",
		consts.HeaderXAPIKey:    "my-api-key",
		consts.HeaderXSign:      "somesign",
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for future timestamp, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: API key validation via DB
// ---------------------------------------------------------------------------

func TestSignature_InvalidAPIKeyInDB(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{findErr: errors.New("not found")}
	mw := newMiddleware(repo, 10)

	now := time.Now().Unix()
	timeStr := strconv.FormatInt(now, 10)
	apiKey := "bad-key"
	reqID := "req-123"
	sign := makeSign(timeStr, apiKey, reqID)

	headers := map[string]string{
		consts.HeaderXTimeUnix:  timeStr,
		consts.HeaderXRequestID: reqID,
		consts.HeaderXAPIKey:    apiKey,
		consts.HeaderXSign:      sign,
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	if w.Code != 498 {
		t.Errorf("expected 498 for invalid API key in DB, got %d", w.Code)
	}
}

func TestSignature_InactiveAPIKey(t *testing.T) {
	t.Parallel()

	repo := &mockReadRepo{
		apiKeyView: &domain.IntegrationAPIKeyView{
			ID:            uuid.New(),
			IntegrationID: domain.NewIntegrationID(),
			Key:           "my-key",
			Active:        false, // inactive
		},
	}
	mw := newMiddleware(repo, 10)

	now := time.Now().Unix()
	timeStr := strconv.FormatInt(now, 10)
	apiKey := "my-key"
	reqID := "req-123"
	sign := makeSign(timeStr, apiKey, reqID)

	headers := map[string]string{
		consts.HeaderXTimeUnix:  timeStr,
		consts.HeaderXRequestID: reqID,
		consts.HeaderXAPIKey:    apiKey,
		consts.HeaderXSign:      sign,
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	// The ValidateAPIKeyHandler returns domain.ErrAPIKeyInactive for inactive keys
	if w.Code != 498 {
		t.Errorf("expected 498 for inactive API key, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: Signature validation
// ---------------------------------------------------------------------------

func TestSignature_InvalidSign(t *testing.T) {
	t.Parallel()

	integrationID := domain.NewIntegrationID()
	apiKeyID := uuid.New()
	repo := &mockReadRepo{
		apiKeyView: &domain.IntegrationAPIKeyView{
			ID:            apiKeyID,
			IntegrationID: integrationID,
			Key:           "valid-key",
			Active:        true,
		},
	}
	mw := newMiddleware(repo, 10)

	now := time.Now().Unix()
	timeStr := strconv.FormatInt(now, 10)
	apiKey := "valid-key"
	reqID := "req-123"

	headers := map[string]string{
		consts.HeaderXTimeUnix:  timeStr,
		consts.HeaderXRequestID: reqID,
		consts.HeaderXAPIKey:    apiKey,
		consts.HeaderXSign:      "wrong-signature",
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	if w.Code != 498 {
		t.Errorf("expected 498 for invalid sign, got %d", w.Code)
	}
}

func TestSignature_ValidRequest(t *testing.T) {
	t.Parallel()

	integrationID := domain.NewIntegrationID()
	apiKeyID := uuid.New()
	repo := &mockReadRepo{
		apiKeyView: &domain.IntegrationAPIKeyView{
			ID:            apiKeyID,
			IntegrationID: integrationID,
			Key:           "valid-key",
			Active:        true,
		},
	}
	mw := newMiddleware(repo, 10)

	now := time.Now().Unix()
	timeStr := strconv.FormatInt(now, 10)
	apiKey := "valid-key"
	reqID := "req-456"
	sign := makeSign(timeStr, apiKey, reqID)

	var capturedIntegrationID, capturedAPIKeyID any
	router := gin.New()
	router.Use(mw.Validate())
	router.GET("/api/v1/test", func(c *gin.Context) {
		capturedIntegrationID, _ = c.Get(consts.CtxIntegrationID)
		capturedAPIKeyID, _ = c.Get(consts.CtxAPIKeyID)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	req.Header.Set(consts.HeaderXTimeUnix, timeStr)
	req.Header.Set(consts.HeaderXRequestID, reqID)
	req.Header.Set(consts.HeaderXAPIKey, apiKey)
	req.Header.Set(consts.HeaderXSign, sign)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for valid request, got %d", w.Code)
	}
	if capturedIntegrationID != integrationID {
		t.Errorf("expected integration ID %v, got %v", integrationID, capturedIntegrationID)
	}
	if capturedAPIKeyID != apiKeyID {
		t.Errorf("expected API key ID %v, got %v", apiKeyID, capturedAPIKeyID)
	}
}

// ---------------------------------------------------------------------------
// Tests: Default expire time
// ---------------------------------------------------------------------------

func TestSignature_DefaultExpireTime(t *testing.T) {
	t.Parallel()

	integrationID := domain.NewIntegrationID()
	apiKeyID := uuid.New()
	repo := &mockReadRepo{
		apiKeyView: &domain.IntegrationAPIKeyView{
			ID:            apiKeyID,
			IntegrationID: integrationID,
			Key:           "valid-key",
			Active:        true,
		},
	}
	// expireTime = 0 triggers default of 10 seconds
	mw := newMiddleware(repo, 0)

	now := time.Now().Unix()
	timeStr := strconv.FormatInt(now, 10)
	apiKey := "valid-key"
	reqID := "req-789"
	sign := makeSign(timeStr, apiKey, reqID)

	headers := map[string]string{
		consts.HeaderXTimeUnix:  timeStr,
		consts.HeaderXRequestID: reqID,
		consts.HeaderXAPIKey:    apiKey,
		consts.HeaderXSign:      sign,
	}
	w := performRequest(mw.Validate(), "GET", "/api/v1/test", headers)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with default expire time, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: SHA256 signature computation
// ---------------------------------------------------------------------------

func TestMakeSign_Deterministic(t *testing.T) {
	t.Parallel()

	timeStr := "1711234567"
	apiKey := "test-key-abc"
	reqID := "req-001"

	sign1 := makeSign(timeStr, apiKey, reqID)
	sign2 := makeSign(timeStr, apiKey, reqID)

	if sign1 != sign2 {
		t.Error("expected SHA256 sign to be deterministic")
	}

	// Verify against known computation
	data := timeStr + apiKey + reqID
	hash := sha256.Sum256([]byte(data))
	expected := hex.EncodeToString(hash[:])
	if sign1 != expected {
		t.Errorf("sign mismatch: got %q, expected %q", sign1, expected)
	}
}

func TestMakeSign_DifferentInputs(t *testing.T) {
	t.Parallel()

	sign1 := makeSign("1111111111", "key-a", "req-1")
	sign2 := makeSign("2222222222", "key-b", "req-2")
	if sign1 == sign2 {
		t.Error("expected different signs for different inputs")
	}
}

// ---------------------------------------------------------------------------
// Tests: API key from query parameter fallback
// ---------------------------------------------------------------------------

func TestSignature_APIKeyFromQueryParam(t *testing.T) {
	t.Parallel()

	integrationID := domain.NewIntegrationID()
	apiKeyID := uuid.New()
	repo := &mockReadRepo{
		apiKeyView: &domain.IntegrationAPIKeyView{
			ID:            apiKeyID,
			IntegrationID: integrationID,
			Key:           "query-key",
			Active:        true,
		},
	}
	mw := newMiddleware(repo, 10)

	now := time.Now().Unix()
	timeStr := strconv.FormatInt(now, 10)
	apiKey := "query-key"
	reqID := "req-qp"
	sign := makeSign(timeStr, apiKey, reqID)

	// No X-API-KEY header, pass via query param
	url := fmt.Sprintf("/api/v1/test?api_key=%s", apiKey)
	headers := map[string]string{
		consts.HeaderXTimeUnix:  timeStr,
		consts.HeaderXRequestID: reqID,
		consts.HeaderXSign:      sign,
	}
	w := performRequest(mw.Validate(), "GET", url, headers)
	if w.Code != http.StatusOK {
		body := w.Body.String()
		var resp map[string]any
		_ = json.Unmarshal([]byte(body), &resp)
		t.Errorf("expected 200 for API key from query param, got %d; body: %s", w.Code, body)
	}
}
