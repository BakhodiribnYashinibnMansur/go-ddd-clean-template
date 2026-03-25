package dashboard

import (
	"gct/internal/dashboard/application/query"
	"gct/internal/dashboard/infrastructure/postgres"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all query handlers for the Dashboard read-only BC.
type BoundedContext struct {
	// Queries
	GetStats *query.GetStatsHandler
}

// NewBoundedContext creates a fully wired Dashboard bounded context (read-only).
func NewBoundedContext(pool *pgxpool.Pool, l logger.Log) *BoundedContext {
	readRepo := postgres.NewDashboardReadRepo(pool)

	return &BoundedContext{
		GetStats: query.NewGetStatsHandler(readRepo, l),
	}
}
