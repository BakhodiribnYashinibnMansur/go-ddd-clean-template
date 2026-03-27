package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/dashboard"
	appdto "gct/internal/dashboard/application"
	"gct/internal/dashboard/application/query"

	"github.com/gin-gonic/gin"
)

// --- Mocks ---

type mockReadRepo struct {
	stats *appdto.DashboardStatsView
}

func (m *mockReadRepo) GetStats(_ context.Context) (*appdto.DashboardStatsView, error) {
	return m.stats, nil
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

	bc := &dashboard.BoundedContext{
		GetStats: query.NewGetStatsHandler(readRepo, l),
	}

	r := gin.New()
	h := NewHandler(bc, l)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

// --- Tests ---

func TestHandler_GetStats_Success(t *testing.T) {
	readRepo := &mockReadRepo{
		stats: &appdto.DashboardStatsView{
			TotalUsers:        150,
			ActiveSessions:    42,
			AuditLogsToday:    87,
			SystemErrorsCount: 3,
			TotalFeatureFlags: 12,
			TotalWebhooks:     5,
			TotalJobs:         8,
		},
	}
	router := setupRouter(readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dashboard/stats", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetStats_EmptyStats(t *testing.T) {
	readRepo := &mockReadRepo{
		stats: &appdto.DashboardStatsView{},
	}
	router := setupRouter(readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dashboard/stats", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
