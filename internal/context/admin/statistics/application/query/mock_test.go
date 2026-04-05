package query

import (
	"context"

	appdto "gct/internal/context/admin/statistics/application"
)

// mockStatisticsReadRepository is a shared stub used across handler tests.
type mockStatisticsReadRepository struct {
	overview    *appdto.OverviewView
	userStats   *appdto.UserStatsView
	sessionStats *appdto.SessionStatsView
	errorStats  *appdto.ErrorStatsView
	auditStats  *appdto.AuditStatsView
	securityStats *appdto.SecurityStatsView
	ffStats     *appdto.FeatureFlagStatsView
	contentStats *appdto.ContentStatsView
	integrationStats *appdto.IntegrationStatsView
	err         error
}

func (m *mockStatisticsReadRepository) GetOverview(_ context.Context) (*appdto.OverviewView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.overview, nil
}
func (m *mockStatisticsReadRepository) GetUserStats(_ context.Context) (*appdto.UserStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.userStats, nil
}
func (m *mockStatisticsReadRepository) GetSessionStats(_ context.Context) (*appdto.SessionStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sessionStats, nil
}
func (m *mockStatisticsReadRepository) GetErrorStats(_ context.Context) (*appdto.ErrorStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.errorStats, nil
}
func (m *mockStatisticsReadRepository) GetAuditStats(_ context.Context) (*appdto.AuditStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.auditStats, nil
}
func (m *mockStatisticsReadRepository) GetSecurityStats(_ context.Context) (*appdto.SecurityStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.securityStats, nil
}
func (m *mockStatisticsReadRepository) GetFeatureFlagStats(_ context.Context) (*appdto.FeatureFlagStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ffStats, nil
}
func (m *mockStatisticsReadRepository) GetContentStats(_ context.Context) (*appdto.ContentStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.contentStats, nil
}
func (m *mockStatisticsReadRepository) GetIntegrationStats(_ context.Context) (*appdto.IntegrationStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.integrationStats, nil
}
