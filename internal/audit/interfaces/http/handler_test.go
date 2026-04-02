package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/audit"
	"gct/internal/audit/application/query"
	"gct/internal/audit/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	auditLogs     []*domain.AuditLogView
	auditTotal    int64
	endpointViews []*domain.EndpointHistoryView
	endpointTotal int64
}

func (m *mockReadRepo) ListAuditLogs(_ context.Context, _ domain.AuditLogFilter) ([]*domain.AuditLogView, int64, error) {
	return m.auditLogs, m.auditTotal, nil
}

func (m *mockReadRepo) ListEndpointHistory(_ context.Context, _ domain.EndpointHistoryFilter) ([]*domain.EndpointHistoryView, int64, error) {
	return m.endpointViews, m.endpointTotal, nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                          {}
func (m *mockLogger) Debugf(template string, args ...any)                        {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Info(args ...any)                                           {}
func (m *mockLogger) Infof(template string, args ...any)                         {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Warn(args ...any)                                           {}
func (m *mockLogger) Warnf(template string, args ...any)                         {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Error(args ...any)                                          {}
func (m *mockLogger) Errorf(template string, args ...any)                        {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Fatal(args ...any)                                          {}
func (m *mockLogger) Fatalf(template string, args ...any)                        {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)               {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)                {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)                {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)               {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)               {}

// --- Helpers ---

func setupRouter(readRepo *mockReadRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)

	l := &mockLogger{}

	bc := &audit.BoundedContext{
		ListAuditLogs:       query.NewListAuditLogsHandler(readRepo, l),
		ListEndpointHistory: query.NewListEndpointHistoryHandler(readRepo, l),
	}

	r := gin.New()
	h := NewHandler(bc, l)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

// --- Tests ---

func TestHandler_ListAuditLogs_Success(t *testing.T) {
	userID := uuid.New()
	readRepo := &mockReadRepo{
		auditLogs: []*domain.AuditLogView{
			{ID: uuid.New(), UserID: &userID, Action: domain.AuditActionLogin, Success: true, CreatedAt: time.Now()},
		},
		auditTotal: 1,
	}
	router := setupRouter(readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/audit-logs?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_ListEndpointHistory_Success(t *testing.T) {
	readRepo := &mockReadRepo{
		endpointViews: []*domain.EndpointHistoryView{
			{ID: uuid.New(), Endpoint: "/api/v1/users", Method: "GET", StatusCode: 200, Latency: 15, CreatedAt: time.Now()},
		},
		endpointTotal: 1,
	}
	router := setupRouter(readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/endpoint-history?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_ListAuditLogs_WithPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		auditLogs: []*domain.AuditLogView{
			{ID: uuid.New(), Action: domain.AuditActionLogin, Success: true, CreatedAt: time.Now()},
		},
		auditTotal: 50,
	}
	router := setupRouter(readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/audit-logs?limit=1&offset=10", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_ListEndpointHistory_WithPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		endpointViews: []*domain.EndpointHistoryView{
			{ID: uuid.New(), Endpoint: "/api/v1/jobs", Method: "POST", StatusCode: 201, Latency: 42, CreatedAt: time.Now()},
		},
		endpointTotal: 100,
	}
	router := setupRouter(readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/endpoint-history?limit=5&offset=20", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
