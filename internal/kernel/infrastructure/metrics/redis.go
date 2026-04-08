package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/redis/go-redis/v9"
)

// RegisterRedisPoolMetrics registers observable gauges for Redis connection pool statistics.
// Metrics are collected on each Prometheus scrape via OTel callbacks.
func RegisterRedisPoolMetrics(client *redis.Client, serviceName string) error {
	meter := otel.Meter(serviceName + "/redis")

	totalConns, err := meter.Int64ObservableGauge("redis_pool_total_conns",
		metric.WithDescription("Total number of connections in the Redis pool"),
	)
	if err != nil {
		return fmt.Errorf("metrics.redis.total_conns: %w", err)
	}

	idleConns, err := meter.Int64ObservableGauge("redis_pool_idle_conns",
		metric.WithDescription("Number of idle connections in the Redis pool"),
	)
	if err != nil {
		return fmt.Errorf("metrics.redis.idle_conns: %w", err)
	}

	staleConns, err := meter.Int64ObservableGauge("redis_pool_stale_conns",
		metric.WithDescription("Number of stale connections removed from the Redis pool"),
	)
	if err != nil {
		return fmt.Errorf("metrics.redis.stale_conns: %w", err)
	}

	hits, err := meter.Int64ObservableCounter("redis_pool_hits",
		metric.WithDescription("Number of times a free connection was found in the Redis pool"),
	)
	if err != nil {
		return fmt.Errorf("metrics.redis.hits: %w", err)
	}

	misses, err := meter.Int64ObservableCounter("redis_pool_misses",
		metric.WithDescription("Number of times a free connection was NOT found in the Redis pool"),
	)
	if err != nil {
		return fmt.Errorf("metrics.redis.misses: %w", err)
	}

	if _, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		stat := client.PoolStats()
		o.ObserveInt64(totalConns, int64(stat.TotalConns))
		o.ObserveInt64(idleConns, int64(stat.IdleConns))
		o.ObserveInt64(staleConns, int64(stat.StaleConns))
		o.ObserveInt64(hits, int64(stat.Hits))
		o.ObserveInt64(misses, int64(stat.Misses))
		return nil
	}, totalConns, idleConns, staleConns, hits, misses); err != nil {
		return fmt.Errorf("metrics.redis.register_callback: %w", err)
	}

	return nil
}
