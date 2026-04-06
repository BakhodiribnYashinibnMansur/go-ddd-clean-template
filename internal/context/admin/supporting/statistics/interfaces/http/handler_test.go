package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/context/admin/supporting/statistics"
	"gct/internal/context/admin/supporting/statistics/application/dto"
	"gct/internal/context/admin/supporting/statistics/application/query"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

type mockReadRepo struct{}

func (m *mockReadRepo) GetOverview(_ context.Context) (*dto.OverviewView, error) {
	return &dto.OverviewView{TotalUsers: 1}, nil
}
func (m *mockReadRepo) GetUserStats(_ context.Context) (*dto.UserStatsView, error) {
	return &dto.UserStatsView{Total: 1, ByRole: map[string]int64{}}, nil
}
func (m *mockReadRepo) GetSessionStats(_ context.Context) (*dto.SessionStatsView, error) {
	return &dto.SessionStatsView{}, nil
}
func (m *mockReadRepo) GetErrorStats(_ context.Context) (*dto.ErrorStatsView, error) {
	return &dto.ErrorStatsView{}, nil
}
func (m *mockReadRepo) GetAuditStats(_ context.Context) (*dto.AuditStatsView, error) {
	return &dto.AuditStatsView{}, nil
}
func (m *mockReadRepo) GetSecurityStats(_ context.Context) (*dto.SecurityStatsView, error) {
	return &dto.SecurityStatsView{}, nil
}
func (m *mockReadRepo) GetFeatureFlagStats(_ context.Context) (*dto.FeatureFlagStatsView, error) {
	return &dto.FeatureFlagStatsView{}, nil
}
func (m *mockReadRepo) GetContentStats(_ context.Context) (*dto.ContentStatsView, error) {
	return &dto.ContentStatsView{}, nil
}
func (m *mockReadRepo) GetIntegrationStats(_ context.Context) (*dto.IntegrationStatsView, error) {
	return &dto.IntegrationStatsView{}, nil
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
	t.Parallel()

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
