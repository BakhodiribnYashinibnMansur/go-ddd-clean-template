package metric

import (
	"gct/internal/context/ops/generic/metric/application/command"
	"gct/internal/context/ops/generic/metric/application/query"
	"gct/internal/context/ops/generic/metric/infrastructure/postgres"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/outbox"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the Metric BC.
type BoundedContext struct {
	// Commands
	RecordMetric *command.RecordMetricHandler

	// Queries
	ListMetrics *query.ListMetricsHandler
}

// NewBoundedContext creates a fully wired Metric bounded context.
func NewBoundedContext(pool *pgxpool.Pool, committer *outbox.EventCommitter, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewMetricWriteRepo(pool)
	readRepo := postgres.NewMetricReadRepo(pool)

	return &BoundedContext{
		RecordMetric: command.NewRecordMetricHandler(writeRepo, committer, l),
		ListMetrics:  query.NewListMetricsHandler(readRepo, l),
	}
}
