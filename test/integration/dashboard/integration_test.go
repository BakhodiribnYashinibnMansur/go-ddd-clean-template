package dashboard

import (
	"context"
	"testing"

	"gct/internal/dashboard"
	"gct/internal/dashboard/application/query"
	"gct/internal/shared/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *dashboard.BoundedContext {
	t.Helper()
	l := logger.New("error")
	return dashboard.NewBoundedContext(setup.TestPG.Pool, l)
}

func TestIntegration_GetStats_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	stats, err := bc.GetStats.Handle(ctx, query.GetStatsQuery{})
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	if stats.TotalUsers != 0 {
		t.Errorf("expected TotalUsers 0, got %d", stats.TotalUsers)
	}
	if stats.ActiveSessions != 0 {
		t.Errorf("expected ActiveSessions 0, got %d", stats.ActiveSessions)
	}
	if stats.AuditLogsToday != 0 {
		t.Errorf("expected AuditLogsToday 0, got %d", stats.AuditLogsToday)
	}
	if stats.SystemErrorsCount != 0 {
		t.Errorf("expected SystemErrorsCount 0, got %d", stats.SystemErrorsCount)
	}
	if stats.TotalFeatureFlags != 0 {
		t.Errorf("expected TotalFeatureFlags 0, got %d", stats.TotalFeatureFlags)
	}


}
