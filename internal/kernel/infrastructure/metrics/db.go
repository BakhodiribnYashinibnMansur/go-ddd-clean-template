package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterPoolMetrics registers observable gauges for pgxpool statistics.
// Metrics are collected on each Prometheus scrape via OTel callbacks.
func RegisterPoolMetrics(pool *pgxpool.Pool, serviceName string) error {
	meter := otel.Meter(serviceName + "/db")

	totalConns, err := meter.Int64ObservableGauge("db_pool_total_conns",
		metric.WithDescription("Total number of connections in the pool"),
	)
	if err != nil {
		return fmt.Errorf("metrics.db.total_conns: %w", err)
	}

	idleConns, err := meter.Int64ObservableGauge("db_pool_idle_conns",
		metric.WithDescription("Number of idle connections in the pool"),
	)
	if err != nil {
		return fmt.Errorf("metrics.db.idle_conns: %w", err)
	}

	acquiredConns, err := meter.Int64ObservableGauge("db_pool_acquired_conns",
		metric.WithDescription("Number of currently acquired connections"),
	)
	if err != nil {
		return fmt.Errorf("metrics.db.acquired_conns: %w", err)
	}

	maxConns, err := meter.Int64ObservableGauge("db_pool_max_conns",
		metric.WithDescription("Maximum number of connections allowed"),
	)
	if err != nil {
		return fmt.Errorf("metrics.db.max_conns: %w", err)
	}

	if _, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		stat := pool.Stat()
		o.ObserveInt64(totalConns, int64(stat.TotalConns()))
		o.ObserveInt64(idleConns, int64(stat.IdleConns()))
		o.ObserveInt64(acquiredConns, int64(stat.AcquiredConns()))
		o.ObserveInt64(maxConns, int64(stat.MaxConns()))
		return nil
	}, totalConns, idleConns, acquiredConns, maxConns); err != nil {
		return fmt.Errorf("metrics.db.register_callback: %w", err)
	}

	return nil
}
