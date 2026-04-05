package statistics

import (
	"gct/internal/context/admin/statistics/application/query"
	"gct/internal/context/admin/statistics/infrastructure/postgres"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all query handlers for the Statistics read-only BC.
type BoundedContext struct {
	GetOverview          *query.GetOverviewHandler
	GetUserStats         *query.GetUserStatsHandler
	GetSessionStats      *query.GetSessionStatsHandler
	GetErrorStats        *query.GetErrorStatsHandler
	GetAuditStats        *query.GetAuditStatsHandler
	GetSecurityStats     *query.GetSecurityStatsHandler
	GetFeatureFlagStats  *query.GetFeatureFlagStatsHandler
	GetContentStats      *query.GetContentStatsHandler
	GetIntegrationStats  *query.GetIntegrationStatsHandler
}

// NewBoundedContext creates a fully wired Statistics bounded context (read-only).
func NewBoundedContext(pool *pgxpool.Pool, l logger.Log) *BoundedContext {
	readRepo := postgres.NewStatisticsReadRepo(pool)

	return &BoundedContext{
		GetOverview:         query.NewGetOverviewHandler(readRepo, l),
		GetUserStats:        query.NewGetUserStatsHandler(readRepo, l),
		GetSessionStats:     query.NewGetSessionStatsHandler(readRepo, l),
		GetErrorStats:       query.NewGetErrorStatsHandler(readRepo, l),
		GetAuditStats:       query.NewGetAuditStatsHandler(readRepo, l),
		GetSecurityStats:    query.NewGetSecurityStatsHandler(readRepo, l),
		GetFeatureFlagStats: query.NewGetFeatureFlagStatsHandler(readRepo, l),
		GetContentStats:     query.NewGetContentStatsHandler(readRepo, l),
		GetIntegrationStats: query.NewGetIntegrationStatsHandler(readRepo, l),
	}
}
