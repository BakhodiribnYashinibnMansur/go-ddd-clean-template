// Package postgres implements postgres connection.
package postgres

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

// Option defines a functional option for configuring Postgres.
type Option func(*pgxpool.Config)

// WithMaxConns sets the maximum number of connections in the pool.
func WithMaxConns(maxConns int32) Option {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConns = maxConns
	}
}

// WithMinConns sets the minimum number of idle connections.
func WithMinConns(minConns int32) Option {
	return func(cfg *pgxpool.Config) {
		cfg.MinConns = minConns
	}
}

// WithMaxConnLifetime sets the maximum lifetime of a connection.
func WithMaxConnLifetime(d time.Duration) Option {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConnLifetime = d
	}
}

// WithMaxConnIdleTime sets the maximum idle time before connection is closed.
func WithMaxConnIdleTime(d time.Duration) Option {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConnIdleTime = d
	}
}

// WithHealthCheckPeriod sets the health check interval.
func WithHealthCheckPeriod(d time.Duration) Option {
	return func(cfg *pgxpool.Config) {
		cfg.HealthCheckPeriod = d
	}
}

// WithConnectTimeout sets the connection establishment timeout.
func WithConnectTimeout(d time.Duration) Option {
	return func(cfg *pgxpool.Config) {
		cfg.ConnConfig.ConnectTimeout = d
	}
}

// WithTraceLogLevel sets the trace log level for pgx.
func WithTraceLogLevel(level tracelog.LogLevel) Option {
	return func(cfg *pgxpool.Config) {
		if tracer, ok := cfg.ConnConfig.Tracer.(*tracelog.TraceLog); ok {
			tracer.LogLevel = level
		}
	}
}

// WithStatementTimeout sets the statement timeout for queries.
func WithStatementTimeout(d time.Duration) Option {
	return func(cfg *pgxpool.Config) {
		if cfg.ConnConfig.RuntimeParams == nil {
			cfg.ConnConfig.RuntimeParams = make(map[string]string)
		}
		cfg.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprintf("%dms", d.Milliseconds())
	}
}

// WithApplicationName sets the application name in PostgreSQL.
func WithApplicationName(name string) Option {
	return func(cfg *pgxpool.Config) {
		if cfg.ConnConfig.RuntimeParams == nil {
			cfg.ConnConfig.RuntimeParams = make(map[string]string)
		}
		cfg.ConnConfig.RuntimeParams["application_name"] = name
	}
}
