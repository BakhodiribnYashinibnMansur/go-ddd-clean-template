package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/context/admin/statistics"
	appdto "gct/internal/context/admin/statistics/application"
	"gct/internal/context/admin/statistics/application/query"
	"gct/internal/platform/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

type mockReadRepo struct{}

func (m *mockReadRepo) GetOverview(_ context.Context) (*appdto.OverviewView, error) {
	return &appdto.OverviewView{TotalUsers: 1}, nil
}
func (m *mockReadRepo) GetUserStats(_ context.Context) (*appdto.UserStatsView, error) {
	return &appdto.UserStatsView{Total: 1, ByRole: map[string]int64{}}, nil
}
func (m *mockReadRepo) GetSessionStats(_ context.Context) (*appdto.SessionStatsView, error) {
	return &appdto.SessionStatsView{}, nil
}
func (m *mockReadRepo) GetErrorStats(_ context.Context) (*appdto.ErrorStatsView, error) {
	return &appdto.ErrorStatsView{}, nil
}
func (m *mockReadRepo) GetAuditStats(_ context.Context) (*appdto.AuditStatsView, error) {
	return &appdto.AuditStatsView{}, nil
}
func (m *mockReadRepo) GetSecurityStats(_ context.Context) (*appdto.SecurityStatsView, error) {
	return &appdto.SecurityStatsView{}, nil
}
func (m *mockReadRepo) GetFeatureFlagStats(_ context.Context) (*appdto.FeatureFlagStatsView, error) {
	return &appdto.FeatureFlagStatsView{}, nil
}
func (m *mockReadRepo) GetContentStats(_ context.Context) (*appdto.ContentStatsView, error) {
	return &appdto.ContentStatsView{}, nil
}
func (m *mockReadRepo) GetIntegrationStats(_ context.Context) (*appdto.IntegrationStatsView, error) {
	return &appdto.IntegrationStatsView{}, nil
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	l := logger.Noop()
	repo := &mockReadRepo{}
	bc := &statistics.BoundedContext{
		GetOverview:         query.NewGetOverviewHandler(repo, l),
		GetUserStats:        query.NewGetUserStatsHandler(repo, l),
		GetSessionStats:     query.NewGetSessionStatsHandler(repo, l),
		GetErrorStats:       query.NewGetErrorStatsHandler(repo, l),
		GetAuditStats:       query.NewGetAuditStatsHandler(repo, l),
		GetSecurityStats:    query.NewGetSecurityStatsHandler(repo, l),
		GetFeatureFlagStats: query.NewGetFeatureFlagStatsHandler(repo, l),
		GetContentStats:     query.NewGetContentStatsHandler(repo, l),
		GetIntegrationStats: query.NewGetIntegrationStatsHandler(repo, l),
	}
	r := gin.New()
	h := NewHandler(bc, l)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

func TestHandler_AllEndpoints_ReturnOK(t *testing.T) {
	router := setupRouter()
	paths := []string{
		"/api/v1/statistics/overview",
		"/api/v1/statistics/users",
		"/api/v1/statistics/sessions",
		"/api/v1/statistics/errors",
		"/api/v1/statistics/audit",
		"/api/v1/statistics/security",
		"/api/v1/statistics/feature-flags",
		"/api/v1/statistics/content",
		"/api/v1/statistics/integrations",
	}
	for _, p := range paths {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", p, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d: %s", p, w.Code, w.Body.String())
		}
	}
}
