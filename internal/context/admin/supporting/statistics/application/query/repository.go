package query

import (
	"context"

	appdto "gct/internal/context/admin/supporting/statistics/application"
)

// StatisticsReadRepository defines the read-side persistence contract for all statistics views.
type StatisticsReadRepository interface {
	GetOverview(ctx context.Context) (*appdto.OverviewView, error)
	GetUserStats(ctx context.Context) (*appdto.UserStatsView, error)
	GetSessionStats(ctx context.Context) (*appdto.SessionStatsView, error)
	GetErrorStats(ctx context.Context) (*appdto.ErrorStatsView, error)
	GetAuditStats(ctx context.Context) (*appdto.AuditStatsView, error)
	GetSecurityStats(ctx context.Context) (*appdto.SecurityStatsView, error)
	GetFeatureFlagStats(ctx context.Context) (*appdto.FeatureFlagStatsView, error)
	GetContentStats(ctx context.Context) (*appdto.ContentStatsView, error)
	GetIntegrationStats(ctx context.Context) (*appdto.IntegrationStatsView, error)
}
