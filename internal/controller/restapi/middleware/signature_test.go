package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/config"
	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/internal/usecase/integration"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIntegrationUC mocks the integration use case for signature tests.
type MockIntegrationUC struct {
	mock.Mock
}

func (m *MockIntegrationUC) CreateIntegration(ctx context.Context, req domain.CreateIntegrationRequest) (*domain.Integration, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.Integration), args.Error(1)
}

func (m *MockIntegrationUC) GetIntegration(ctx context.Context, id uuid.UUID) (*domain.IntegrationWithKeys, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.IntegrationWithKeys), args.Error(1)
}

func (m *MockIntegrationUC) ListIntegrations(ctx context.Context, filter domain.IntegrationFilter) ([]domain.Integration, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Integration), args.Get(1).(int64), args.Error(2)
}

func (m *MockIntegrationUC) UpdateIntegration(ctx context.Context, id uuid.UUID, req domain.UpdateIntegrationRequest) (*domain.Integration, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*domain.Integration), args.Error(1)
}

func (m *MockIntegrationUC) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIntegrationUC) ToggleIntegration(ctx context.Context, id uuid.UUID) (*domain.Integration, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Integration), args.Error(1)
}

func (m *MockIntegrationUC) CreateAPIKey(ctx context.Context, req domain.CreateAPIKeyRequest) (*domain.APIKey, string, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.APIKey), args.String(1), args.Error(2)
}

func (m *MockIntegrationUC) GetAPIKey(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func (m *MockIntegrationUC) ListAPIKeys(ctx context.Context, integrationID uuid.UUID) ([]domain.APIKey, error) {
	args := m.Called(ctx, integrationID)
	return args.Get(0).([]domain.APIKey), args.Error(1)
}

func (m *MockIntegrationUC) ValidateAPIKey(ctx context.Context, key string) (*domain.APIKey, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func (m *MockIntegrationUC) RevokeAPIKey(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIntegrationUC) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIntegrationUC) InitCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockIntegrationUC) InvalidateCache(ctx context.Context, table string) error {
	args := m.Called(ctx, table)
	return args.Error(0)
}

// Ensure mock implements the interface.
var _ integration.UseCaseI = (*MockIntegrationUC)(nil)

func buildSignature(timeUnix, apiKey, requestID string) string {
	data := timeUnix + apiKey + requestID
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func newSignatureTestCfg() *config.Config {
	return &config.Config{
		APIKeys: config.APIKeys{
			SignExpireTime: 10,
		},
	}
}

func TestSignatureMiddleware_SkipsAdminRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/admin/dashboard", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSignatureMiddleware_SkipsHealthRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSignatureMiddleware_SkipsDocsRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/docs/swagger", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs/swagger", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSignatureMiddleware_SkipsRootRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSignatureMiddleware_MissingTimeUnixReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "error", body["status"])
}

func TestSignatureMiddleware_MissingRequestIDReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	req.Header.Set(consts.HeaderXTimeUnix, fmt.Sprintf("%d", time.Now().Unix()))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSignatureMiddleware_MissingAPIKeyReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	req.Header.Set(consts.HeaderXTimeUnix, fmt.Sprintf("%d", time.Now().Unix()))
	req.Header.Set(consts.HeaderXRequestID, uuid.New().String())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSignatureMiddleware_MissingSignReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	req.Header.Set(consts.HeaderXTimeUnix, fmt.Sprintf("%d", time.Now().Unix()))
	req.Header.Set(consts.HeaderXRequestID, uuid.New().String())
	req.Header.Set(consts.HeaderXAPIKey, "test-api-key")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSignatureMiddleware_ExpiredTimestamp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Timestamp 1 hour in the past
	expiredTime := fmt.Sprintf("%d", time.Now().Add(-1*time.Hour).Unix())
	apiKey := "test-api-key"
	requestID := uuid.New().String()
	sign := buildSignature(expiredTime, apiKey, requestID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	req.Header.Set(consts.HeaderXTimeUnix, expiredTime)
	req.Header.Set(consts.HeaderXRequestID, requestID)
	req.Header.Set(consts.HeaderXAPIKey, apiKey)
	req.Header.Set(consts.HeaderXSign, sign)
	r.ServeHTTP(w, req)

	// 499 for expired time
	assert.Equal(t, 499, w.Code)
}

func TestSignatureMiddleware_FutureDatedTimestamp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Timestamp 1 hour in the future
	futureTime := fmt.Sprintf("%d", time.Now().Add(1*time.Hour).Unix())
	apiKey := "test-api-key"
	requestID := uuid.New().String()
	sign := buildSignature(futureTime, apiKey, requestID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	req.Header.Set(consts.HeaderXTimeUnix, futureTime)
	req.Header.Set(consts.HeaderXRequestID, requestID)
	req.Header.Set(consts.HeaderXAPIKey, apiKey)
	req.Header.Set(consts.HeaderXSign, sign)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSignatureMiddleware_InvalidTimeFormatReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	uc := &usecase.UseCase{}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	req.Header.Set(consts.HeaderXTimeUnix, "not-a-number")
	req.Header.Set(consts.HeaderXRequestID, uuid.New().String())
	req.Header.Set(consts.HeaderXAPIKey, "test-api-key")
	req.Header.Set(consts.HeaderXSign, "some-sign")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSignatureMiddleware_InvalidAPIKeyReturns498(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	mockIntegration := new(MockIntegrationUC)

	apiKey := "invalid-api-key"
	mockIntegration.On("ValidateAPIKey", mock.Anything, apiKey).
		Return(nil, fmt.Errorf("key not found"))

	uc := &usecase.UseCase{
		Integration: mockIntegration,
	}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	timeUnix := fmt.Sprintf("%d", time.Now().Unix())
	requestID := uuid.New().String()
	sign := buildSignature(timeUnix, apiKey, requestID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	req.Header.Set(consts.HeaderXTimeUnix, timeUnix)
	req.Header.Set(consts.HeaderXRequestID, requestID)
	req.Header.Set(consts.HeaderXAPIKey, apiKey)
	req.Header.Set(consts.HeaderXSign, sign)
	r.ServeHTTP(w, req)

	assert.Equal(t, 498, w.Code)
	mockIntegration.AssertExpectations(t)
}

func TestSignatureMiddleware_InvalidSignatureReturns498(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	mockIntegration := new(MockIntegrationUC)

	apiKey := "valid-api-key"
	integrationID := uuid.New()
	apiKeyID := uuid.New()
	mockIntegration.On("ValidateAPIKey", mock.Anything, apiKey).
		Return(&domain.APIKey{
			ID:            apiKeyID,
			IntegrationID: integrationID,
		}, nil)

	uc := &usecase.UseCase{
		Integration: mockIntegration,
	}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	timeUnix := fmt.Sprintf("%d", time.Now().Unix())
	requestID := uuid.New().String()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	req.Header.Set(consts.HeaderXTimeUnix, timeUnix)
	req.Header.Set(consts.HeaderXRequestID, requestID)
	req.Header.Set(consts.HeaderXAPIKey, apiKey)
	req.Header.Set(consts.HeaderXSign, "wrong-signature-value")
	r.ServeHTTP(w, req)

	assert.Equal(t, 498, w.Code)
	mockIntegration.AssertExpectations(t)
}

func TestSignatureMiddleware_ValidSignaturePasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	mockIntegration := new(MockIntegrationUC)

	apiKey := "valid-api-key"
	integrationID := uuid.New()
	apiKeyID := uuid.New()
	mockIntegration.On("ValidateAPIKey", mock.Anything, apiKey).
		Return(&domain.APIKey{
			ID:            apiKeyID,
			IntegrationID: integrationID,
		}, nil)

	uc := &usecase.UseCase{
		Integration: mockIntegration,
	}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		// Verify context values were set
		ctxIntegrationID, exists := c.Get(consts.CtxIntegrationID)
		assert.True(t, exists)
		assert.Equal(t, integrationID, ctxIntegrationID)

		ctxAPIKeyID, exists := c.Get(consts.CtxAPIKeyID)
		assert.True(t, exists)
		assert.Equal(t, apiKeyID, ctxAPIKeyID)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	timeUnix := fmt.Sprintf("%d", time.Now().Unix())
	requestID := uuid.New().String()
	sign := buildSignature(timeUnix, apiKey, requestID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data", nil)
	req.Header.Set(consts.HeaderXTimeUnix, timeUnix)
	req.Header.Set(consts.HeaderXRequestID, requestID)
	req.Header.Set(consts.HeaderXAPIKey, apiKey)
	req.Header.Set(consts.HeaderXSign, sign)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockIntegration.AssertExpectations(t)
}

func TestSignatureMiddleware_APIKeyFromQueryParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newSignatureTestCfg()
	mockIntegration := new(MockIntegrationUC)

	apiKey := "query-api-key"
	integrationID := uuid.New()
	apiKeyID := uuid.New()
	mockIntegration.On("ValidateAPIKey", mock.Anything, apiKey).
		Return(&domain.APIKey{
			ID:            apiKeyID,
			IntegrationID: integrationID,
		}, nil)

	uc := &usecase.UseCase{
		Integration: mockIntegration,
	}

	r := gin.New()
	r.Use(SignatureMiddleware(cfg, uc))
	r.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	timeUnix := fmt.Sprintf("%d", time.Now().Unix())
	requestID := uuid.New().String()
	sign := buildSignature(timeUnix, apiKey, requestID)

	w := httptest.NewRecorder()
	// API key in query param instead of header
	req := httptest.NewRequest(http.MethodGet, "/api/v1/data?api_key="+apiKey, nil)
	req.Header.Set(consts.HeaderXTimeUnix, timeUnix)
	req.Header.Set(consts.HeaderXRequestID, requestID)
	// No X-API-KEY header
	req.Header.Set(consts.HeaderXSign, sign)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockIntegration.AssertExpectations(t)
}
