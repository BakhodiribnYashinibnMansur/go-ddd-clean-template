package query

import (
	"context"

	"gct/internal/context/admin/supporting/statistics/application/dto"
)

// mockStatisticsReadRepository is a shared stub used across handler tests.
type mockStatisticsReadRepository struct {
	overview    *dto.OverviewView
	userStats   *dto.UserStatsView
	sessionStats *dto.SessionStatsView
	errorStats  *dto.ErrorStatsView
	auditStats  *dto.AuditStatsView
	securityStats *dto.SecurityStatsView
	ffStats     *dto.FeatureFlagStatsView
	contentStats *dto.ContentStatsView
	integrationStats *dto.IntegrationStatsView
	err         error
}

func (m *mockStatisticsReadRepository) GetOverview(_ context.Context) (*dto.OverviewView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.overview, nil
}
func (m *mockStatisticsReadRepository) GetUserStats(_ context.Context) (*dto.UserStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.userStats, nil
}
func (m *mockStatisticsReadRepository) GetSessionStats(_ context.Context) (*dto.SessionStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sessionStats, nil
}
func (m *mockStatisticsReadRepository) GetErrorStats(_ context.Context) (*dto.ErrorStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.errorStats, nil
}
func (m *mockStatisticsReadRepository) GetAuditStats(_ context.Context) (*dto.AuditStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.auditStats, nil
}
func (m *mockStatisticsReadRepository) GetSecurityStats(_ context.Context) (*dto.SecurityStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.securityStats, nil
}
func (m *mockStatisticsReadRepository) GetFeatureFlagStats(_ context.Context) (*dto.FeatureFlagStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ffStats, nil
}
func (m *mockStatisticsReadRepository) GetContentStats(_ context.Context) (*dto.ContentStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.contentStats, nil
}
func (m *mockStatisticsReadRepository) GetIntegrationStats(_ context.Context) (*dto.IntegrationStatsView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.integrationStats, nil
}
