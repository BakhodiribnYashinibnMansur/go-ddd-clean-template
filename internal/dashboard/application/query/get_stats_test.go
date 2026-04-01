package query

import (
	"context"
	"errors"
	"testing"

	appdto "gct/internal/dashboard/application"
)

// --- Mock Read Repository ---

type mockDashboardReadRepository struct {
	stats *appdto.DashboardStatsView
	err   error
}

func (m *mockDashboardReadRepository) GetStats(_ context.Context) (*appdto.DashboardStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

// --- Mock Logger ---

type mockLogger struct{}

func (ml *mockLogger) Debug(_ ...any)                                       {}
func (ml *mockLogger) Debugf(_ string, _ ...any)                            {}
func (ml *mockLogger) Debugw(_ string, _ ...any)                            {}
func (ml *mockLogger) Info(_ ...any)                                        {}
func (ml *mockLogger) Infof(_ string, _ ...any)                             {}
func (ml *mockLogger) Infow(_ string, _ ...any)                             {}
func (ml *mockLogger) Warn(_ ...any)                                        {}
func (ml *mockLogger) Warnf(_ string, _ ...any)                             {}
func (ml *mockLogger) Warnw(_ string, _ ...any)                             {}
func (ml *mockLogger) Error(_ ...any)                                       {}
func (ml *mockLogger) Errorf(_ string, _ ...any)                            {}
func (ml *mockLogger) Errorw(_ string, _ ...any)                            {}
func (ml *mockLogger) Fatal(_ ...any)                                       {}
func (ml *mockLogger) Fatalf(_ string, _ ...any)                            {}
func (ml *mockLogger) Fatalw(_ string, _ ...any)                            {}
func (ml *mockLogger) Debugc(_ context.Context, _ string, _ ...any)         {}
func (ml *mockLogger) Infoc(_ context.Context, _ string, _ ...any)          {}
func (ml *mockLogger) Warnc(_ context.Context, _ string, _ ...any)          {}
func (ml *mockLogger) Errorc(_ context.Context, _ string, _ ...any)         {}
func (ml *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)         {}

// --- Tests ---

func TestGetStatsHandler_Handle(t *testing.T) {
	repo := &mockDashboardReadRepository{
		stats: &appdto.DashboardStatsView{
			TotalUsers:        150,
			ActiveSessions:    42,
			AuditLogsToday:    87,
			SystemErrorsCount: 3,
			TotalFeatureFlags: 12,


		},
	}

	handler := NewGetStatsHandler(repo, &mockLogger{})

	result, err := handler.Handle(context.Background(), GetStatsQuery{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected dashboard stats view, got nil")
	}

	if result.TotalUsers != 150 {
		t.Errorf("expected TotalUsers 150, got %d", result.TotalUsers)
	}

	if result.ActiveSessions != 42 {
		t.Errorf("expected ActiveSessions 42, got %d", result.ActiveSessions)
	}

	if result.AuditLogsToday != 87 {
		t.Errorf("expected AuditLogsToday 87, got %d", result.AuditLogsToday)
	}

	if result.SystemErrorsCount != 3 {
		t.Errorf("expected SystemErrorsCount 3, got %d", result.SystemErrorsCount)
	}

	if result.TotalFeatureFlags != 12 {
		t.Errorf("expected TotalFeatureFlags 12, got %d", result.TotalFeatureFlags)
	}



}

func TestGetStatsHandler_RepoError(t *testing.T) {
	repo := &mockDashboardReadRepository{
		err: errors.New("database connection failed"),
	}

	handler := NewGetStatsHandler(repo, &mockLogger{})

	_, err := handler.Handle(context.Background(), GetStatsQuery{})
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}
