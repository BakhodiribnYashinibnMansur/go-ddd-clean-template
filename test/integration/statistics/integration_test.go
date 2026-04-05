package statistics

import (
	"context"
	"testing"

	"gct/internal/context/admin/supporting/statistics"
	"gct/internal/context/admin/supporting/statistics/application/query"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *statistics.BoundedContext {
	t.Helper()
	l := logger.New("error")
	return statistics.NewBoundedContext(setup.TestPG.Pool, l)
}

func TestIntegration_GetOverview_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	view, err := bc.GetOverview.Handle(ctx, query.GetOverviewQuery{})
	if err != nil {
		t.Fatalf("GetOverview: %v", err)
	}
	if view.TotalUsers != 0 {
		t.Errorf("TotalUsers: got %d", view.TotalUsers)
	}
	if view.ActiveSessions != 0 {
		t.Errorf("ActiveSessions: got %d", view.ActiveSessions)
	}
	if view.AuditLogsToday != 0 {
		t.Errorf("AuditLogsToday: got %d", view.AuditLogsToday)
	}
	if view.SystemErrorsCount != 0 {
		t.Errorf("SystemErrorsCount: got %d", view.SystemErrorsCount)
	}
	if view.TotalFeatureFlags != 0 {
		t.Errorf("TotalFeatureFlags: got %d", view.TotalFeatureFlags)
	}
}

func TestIntegration_GetUserStats_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	view, err := bc.GetUserStats.Handle(context.Background(), query.GetUserStatsQuery{})
	if err != nil {
		t.Fatalf("GetUserStats: %v", err)
	}
	if view.Total != 0 || view.Deleted != 0 {
		t.Errorf("unexpected user stats: %+v", view)
	}
	if view.ByRole == nil {
		t.Error("ByRole map should be initialized")
	}
}

func TestIntegration_GetSessionStats_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	view, err := bc.GetSessionStats.Handle(context.Background(), query.GetSessionStatsQuery{})
	if err != nil {
		t.Fatalf("GetSessionStats: %v", err)
	}
	if view.Active != 0 || view.Expired != 0 || view.Revoked != 0 {
		t.Errorf("unexpected session stats: %+v", view)
	}
}

func TestIntegration_GetErrorStats_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	view, err := bc.GetErrorStats.Handle(context.Background(), query.GetErrorStatsQuery{})
	if err != nil {
		t.Fatalf("GetErrorStats: %v", err)
	}
	if view.Unresolved != 0 || view.Resolved != 0 || view.Last24h != 0 {
		t.Errorf("unexpected error stats: %+v", view)
	}
}

func TestIntegration_GetAuditStats_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	view, err := bc.GetAuditStats.Handle(context.Background(), query.GetAuditStatsQuery{})
	if err != nil {
		t.Fatalf("GetAuditStats: %v", err)
	}
	if view.Today != 0 || view.Last7Days != 0 || view.Total != 0 {
		t.Errorf("unexpected audit stats: %+v", view)
	}
}

func TestIntegration_GetSecurityStats_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	view, err := bc.GetSecurityStats.Handle(context.Background(), query.GetSecurityStatsQuery{})
	if err != nil {
		t.Fatalf("GetSecurityStats: %v", err)
	}
	if view.IPRules != 0 || view.RateLimits != 0 {
		t.Errorf("unexpected security stats: %+v", view)
	}
}

func TestIntegration_GetFeatureFlagStats_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	view, err := bc.GetFeatureFlagStats.Handle(context.Background(), query.GetFeatureFlagStatsQuery{})
	if err != nil {
		t.Fatalf("GetFeatureFlagStats: %v", err)
	}
	if view.Total != 0 || view.Enabled != 0 || view.Disabled != 0 {
		t.Errorf("unexpected feature flag stats: %+v", view)
	}
}

func TestIntegration_GetContentStats_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	view, err := bc.GetContentStats.Handle(context.Background(), query.GetContentStatsQuery{})
	if err != nil {
		t.Fatalf("GetContentStats: %v", err)
	}
	if view.Announcements != 0 || view.Notifications != 0 || view.FileMetadata != 0 || view.Translations != 0 {
		t.Errorf("unexpected content stats: %+v", view)
	}
}

func TestIntegration_GetIntegrationStats_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	view, err := bc.GetIntegrationStats.Handle(context.Background(), query.GetIntegrationStatsQuery{})
	if err != nil {
		t.Fatalf("GetIntegrationStats: %v", err)
	}
	if view.Integrations != 0 || view.APIKeys != 0 {
		t.Errorf("unexpected integration stats: %+v", view)
	}
}
