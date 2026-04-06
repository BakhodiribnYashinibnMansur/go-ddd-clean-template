package query

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/admin/supporting/statistics/application/dto"
	"gct/internal/kernel/infrastructure/logger"
	"github.com/stretchr/testify/require"
)

func TestGetOverviewHandler(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{
		overview: &dto.OverviewView{
			TotalUsers:        10,
			ActiveSessions:    2,
			AuditLogsToday:    5,
			SystemErrorsCount: 1,
			TotalFeatureFlags: 3,
		},
	}
	h := NewGetOverviewHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background(), GetOverviewQuery{})
	require.NoError(t, err)
	if got.TotalUsers != 10 {
		t.Errorf("TotalUsers: got %d", got.TotalUsers)
	}
}

func TestGetOverviewHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{err: errors.New("boom")}
	h := NewGetOverviewHandler(repo, logger.Noop())
	if _, err := h.Handle(context.Background(), GetOverviewQuery{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetUserStatsHandler(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{
		userStats: &dto.UserStatsView{Total: 7, Deleted: 1, ByRole: map[string]int64{"admin": 2}},
	}
	h := NewGetUserStatsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background(), GetUserStatsQuery{})
	require.NoError(t, err)
	if got.Total != 7 || got.ByRole["admin"] != 2 {
		t.Errorf("unexpected view: %+v", got)
	}
}

func TestGetUserStatsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{err: errors.New("boom")}
	h := NewGetUserStatsHandler(repo, logger.Noop())
	if _, err := h.Handle(context.Background(), GetUserStatsQuery{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetSessionStatsHandler(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{
		sessionStats: &dto.SessionStatsView{Active: 3, Expired: 1, Revoked: 2},
	}
	h := NewGetSessionStatsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background(), GetSessionStatsQuery{})
	require.NoError(t, err)
	if got.Active != 3 {
		t.Errorf("Active: got %d", got.Active)
	}
}

func TestGetSessionStatsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{err: errors.New("boom")}
	h := NewGetSessionStatsHandler(repo, logger.Noop())
	if _, err := h.Handle(context.Background(), GetSessionStatsQuery{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetErrorStatsHandler(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{
		errorStats: &dto.ErrorStatsView{Unresolved: 4, Resolved: 6, Last24h: 2},
	}
	h := NewGetErrorStatsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background(), GetErrorStatsQuery{})
	require.NoError(t, err)
	if got.Unresolved != 4 {
		t.Errorf("Unresolved: got %d", got.Unresolved)
	}
}

func TestGetErrorStatsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{err: errors.New("boom")}
	h := NewGetErrorStatsHandler(repo, logger.Noop())
	if _, err := h.Handle(context.Background(), GetErrorStatsQuery{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetAuditStatsHandler(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{
		auditStats: &dto.AuditStatsView{Today: 1, Last7Days: 7, Total: 100},
	}
	h := NewGetAuditStatsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background(), GetAuditStatsQuery{})
	require.NoError(t, err)
	if got.Total != 100 {
		t.Errorf("Total: got %d", got.Total)
	}
}

func TestGetAuditStatsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{err: errors.New("boom")}
	h := NewGetAuditStatsHandler(repo, logger.Noop())
	if _, err := h.Handle(context.Background(), GetAuditStatsQuery{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetSecurityStatsHandler(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{
		securityStats: &dto.SecurityStatsView{IPRules: 5, RateLimits: 3},
	}
	h := NewGetSecurityStatsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background(), GetSecurityStatsQuery{})
	require.NoError(t, err)
	if got.IPRules != 5 {
		t.Errorf("IPRules: got %d", got.IPRules)
	}
}

func TestGetSecurityStatsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{err: errors.New("boom")}
	h := NewGetSecurityStatsHandler(repo, logger.Noop())
	if _, err := h.Handle(context.Background(), GetSecurityStatsQuery{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetFeatureFlagStatsHandler(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{
		ffStats: &dto.FeatureFlagStatsView{Total: 8, Enabled: 5, Disabled: 3},
	}
	h := NewGetFeatureFlagStatsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background(), GetFeatureFlagStatsQuery{})
	require.NoError(t, err)
	if got.Enabled != 5 {
		t.Errorf("Enabled: got %d", got.Enabled)
	}
}

func TestGetFeatureFlagStatsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{err: errors.New("boom")}
	h := NewGetFeatureFlagStatsHandler(repo, logger.Noop())
	if _, err := h.Handle(context.Background(), GetFeatureFlagStatsQuery{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetContentStatsHandler(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{
		contentStats: &dto.ContentStatsView{Announcements: 1, Notifications: 2, FileMetadata: 3, Translations: 4},
	}
	h := NewGetContentStatsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background(), GetContentStatsQuery{})
	require.NoError(t, err)
	if got.Translations != 4 {
		t.Errorf("Translations: got %d", got.Translations)
	}
}

func TestGetContentStatsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{err: errors.New("boom")}
	h := NewGetContentStatsHandler(repo, logger.Noop())
	if _, err := h.Handle(context.Background(), GetContentStatsQuery{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetIntegrationStatsHandler(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{
		integrationStats: &dto.IntegrationStatsView{Integrations: 2, APIKeys: 3},
	}
	h := NewGetIntegrationStatsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background(), GetIntegrationStatsQuery{})
	require.NoError(t, err)
	if got.APIKeys != 3 {
		t.Errorf("APIKeys: got %d", got.APIKeys)
	}
}

func TestGetIntegrationStatsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockStatisticsReadRepository{err: errors.New("boom")}
	h := NewGetIntegrationStatsHandler(repo, logger.Noop())
	if _, err := h.Handle(context.Background(), GetIntegrationStatsQuery{}); err == nil {
		t.Fatal("expected error")
	}
}
