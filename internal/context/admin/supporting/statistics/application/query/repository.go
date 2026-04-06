package query

import (
	"context"

	"gct/internal/context/admin/supporting/statistics/application/dto"
)

// StatisticsReadRepository defines the read-side persistence contract for all statistics views.
type StatisticsReadRepository interface {
	GetOverview(ctx context.Context) (*dto.OverviewView, error)
	GetUserStats(ctx context.Context) (*dto.UserStatsView, error)
	GetSessionStats(ctx context.Context) (*dto.SessionStatsView, error)
	GetErrorStats(ctx context.Context) (*dto.ErrorStatsView, error)
	GetAuditStats(ctx context.Context) (*dto.AuditStatsView, error)
	GetSecurityStats(ctx context.Context) (*dto.SecurityStatsView, error)
	GetFeatureFlagStats(ctx context.Context) (*dto.FeatureFlagStatsView, error)
	GetContentStats(ctx context.Context) (*dto.ContentStatsView, error)
	GetIntegrationStats(ctx context.Context) (*dto.IntegrationStatsView, error)
}
