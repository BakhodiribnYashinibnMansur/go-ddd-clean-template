package errorx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestNewHTTPErrorLogger(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewHTTPErrorLogger(repo, log)

	if logger == nil {
		t.Fatal("expected non-nil HTTPErrorLogger")
	}
	if logger.errorLogger == nil {
		t.Error("expected non-nil inner errorLogger")
	}
	if logger.logger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestExtractHTTPContext_BasicRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	ctx := ExtractHTTPContext(req)

	if ctx.Path != "/api/v1/users" {
		t.Errorf("expected path '/api/v1/users', got %q", ctx.Path)
	}
	if ctx.Method != http.MethodGet {
		t.Errorf("expected method 'GET', got %q", ctx.Method)
	}
}

func TestExtractHTTPContext_WithRequestIDUUID(t *testing.T) {
	reqID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", reqID))

	ctx := ExtractHTTPContext(req)

	if ctx.RequestID == nil {
		t.Fatal("expected non-nil RequestID")
	}
	if *ctx.RequestID != reqID {
		t.Errorf("expected RequestID %s, got %s", reqID, *ctx.RequestID)
	}
}

func TestExtractHTTPContext_WithRequestIDString(t *testing.T) {
	reqID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", reqID.String()))

	ctx := ExtractHTTPContext(req)

	if ctx.RequestID == nil {
		t.Fatal("expected non-nil RequestID")
	}
	if *ctx.RequestID != reqID {
		t.Errorf("expected RequestID %s, got %s", reqID, *ctx.RequestID)
	}
}

func TestExtractHTTPContext_WithUserIDUUID(t *testing.T) {
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))

	ctx := ExtractHTTPContext(req)

	if ctx.UserID == nil {
		t.Fatal("expected non-nil UserID")
	}
	if *ctx.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, *ctx.UserID)
	}
}

func TestExtractHTTPContext_WithUserIDString(t *testing.T) {
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID.String()))

	ctx := ExtractHTTPContext(req)

	if ctx.UserID == nil {
		t.Fatal("expected non-nil UserID")
	}
	if *ctx.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, *ctx.UserID)
	}
}

func TestExtractHTTPContext_NoContextValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	ctx := ExtractHTTPContext(req)

	if ctx.RequestID != nil {
		t.Error("expected nil RequestID when not in context")
	}
	if ctx.UserID != nil {
		t.Error("expected nil UserID when not in context")
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")

	ip := getClientIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("expected IP '10.0.0.1', got %q", ip)
	}
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.2")

	ip := getClientIP(req)
	if ip != "10.0.0.2" {
		t.Errorf("expected IP '10.0.0.2', got %q", ip)
	}
}

func TestGetClientIP_FallbackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	ip := getClientIP(req)
	if ip != "192.168.1.1:1234" {
		t.Errorf("expected IP '192.168.1.1:1234', got %q", ip)
	}
}

func TestGetClientIP_XForwardedForTakesPrecedence(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.Header.Set("X-Real-IP", "10.0.0.2")
	req.RemoteAddr = "192.168.1.1:1234"

	ip := getClientIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("expected X-Forwarded-For to take precedence, got %q", ip)
	}
}

func TestHTTPErrorLogger_LogHTTPError(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewHTTPErrorLogger(repo, log)

	reqID := uuid.New()
	userID := uuid.New()
	httpCtx := HTTPErrorContext{
		RequestID: &reqID,
		UserID:    &userID,
		IPAddress: "127.0.0.1",
		Path:      "/api/test",
		Method:    "POST",
	}

	err := logger.LogHTTPError(context.Background(), "TEST_ERROR", "test message", nil, httpCtx, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.createCallCount != 1 {
		t.Errorf("expected 1 repo.Create call, got %d", repo.createCallCount)
	}
}

func TestHTTPErrorLogger_LogAuthError(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewHTTPErrorLogger(repo, log)

	httpCtx := HTTPErrorContext{
		Path:   "/api/login",
		Method: "POST",
	}

	err := logger.LogAuthError(context.Background(), nil, httpCtx, "john_doe")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.createCallCount != 1 {
		t.Errorf("expected 1 create call, got %d", repo.createCallCount)
	}
}

func TestHTTPErrorLogger_LogValidationError(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewHTTPErrorLogger(repo, log)

	httpCtx := HTTPErrorContext{
		Path:   "/api/register",
		Method: "POST",
	}

	err := logger.LogValidationError(context.Background(), nil, httpCtx, "email", "invalid-email")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHTTPErrorLogger_LogDatabaseError(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewHTTPErrorLogger(repo, log)

	httpCtx := HTTPErrorContext{
		Path:   "/api/users",
		Method: "GET",
	}

	err := logger.LogDatabaseError(context.Background(), nil, httpCtx, "SELECT", "users")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHTTPErrorLogger_LogExternalServiceError(t *testing.T) {
	repo := &mockRepository{}
	log := &mockLogger{}
	logger := NewHTTPErrorLogger(repo, log)

	httpCtx := HTTPErrorContext{
		Path:   "/api/payment",
		Method: "POST",
	}

	err := logger.LogExternalServiceError(context.Background(), nil, httpCtx, "stripe", "/v1/charges")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestExtractHTTPContext_InvalidUUIDString(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "not-a-uuid"))

	ctx := ExtractHTTPContext(req)

	if ctx.RequestID != nil {
		t.Error("expected nil RequestID for invalid UUID string")
	}
}
