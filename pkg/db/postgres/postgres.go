// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"

	"gct/config"
	"gct/pkg/logger"
)

const (
	// ENV PROD
	maxConnsProd = 50 // Maximum number of connections in the pool
	minConnsProd = 10 // Minimum number of idle connections

	// ENV DEV
	maxConnsDev = 8 // Maximum number of connections in the pool
	minConnsDev = 3 // Minimum number of idle connections

	defaultMaxConnLifetime   = 10 * time.Hour   // Maximum lifetime of a connection
	defaultMaxConnIdleTime   = 30 * time.Minute // Maximum idle time before connection is closed
	defaultHealthCheckPeriod = 5 * time.Minute  // Health check interval
	defaultConnectTimeout    = 10 * time.Second // Connection establishment timeout
	defaultPingTimeout       = 5 * time.Second  // Ping verification timeout
)

// Postgres struct wraps pgxpool.Pool and squirrel.Builder.
type Postgres struct {
	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

// New creates a new PostgreSQL connection pool with optimized settings.
// Additional options can be provided to customize the pool configuration.
func New(ctx context.Context, env string, cfg config.Postgres, l logger.Log, opts ...Option) (*Postgres, error) {
	connString := buildConnectionString(cfg)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		l.Errorw("failed to parse connection string", zap.Error(err))
		return nil, fmt.Errorf("parse connection config: %w", err)
	}

	// Apply pool configuration
	applyPoolConfig(env, poolConfig)

	// Set tracer
	setTracer(poolConfig, l)

	// Apply custom options (these override defaults)
	for _, opt := range opts {
		opt(poolConfig)
	}

	// Create connection pool with timeout context
	poolCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(poolCtx, poolConfig)
	if err != nil {
		l.Errorw("failed to create connection pool", zap.Error(err))
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Verify connection with ping
	if err := verifyConnection(ctx, pool, l); err != nil {
		pool.Close()
		return nil, err
	}

	pg := &Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    pool,
	}

	l.Infow("PostgreSQL connected successfully")

	return pg, nil
}

// buildConnectionString constructs the PostgreSQL connection string.
func buildConnectionString(cfg config.Postgres) string {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	if cfg.SSLMode != "" {
		connString += "?sslmode=" + cfg.SSLMode
	}

	return connString
}

// applyPoolConfig applies pool configuration with defaults.
func applyPoolConfig(env string, poolConfig *pgxpool.Config) {
	if env == "production" || env == "PROD" {
		poolConfig.MaxConns = maxConnsProd
		poolConfig.MinConns = minConnsProd
	} else {
		poolConfig.MaxConns = maxConnsDev
		poolConfig.MinConns = minConnsDev
	}

	// Apply default configuration
	poolConfig.MaxConnLifetime = defaultMaxConnLifetime
	poolConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	poolConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout
}

// verifyConnection pings the database to ensure connectivity.
func verifyConnection(ctx context.Context, pool *pgxpool.Pool, l logger.Log) error {
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		l.Errorw("failed to ping database", zap.Error(err))
		return fmt.Errorf("verify database connection: %w", err)
	}

	return nil
}

// Close gracefully closes the database connection pool.
func (p *Postgres) Close() {
	if p != nil && p.Pool != nil {
		p.Pool.Close()
	}
}

// Stats returns current pool statistics.
func (p *Postgres) Stats() *pgxpool.Stat {
	if p == nil || p.Pool == nil {
		return nil
	}
	return p.Pool.Stat()
}

func setTracer(poolConfig *pgxpool.Config, l logger.Log) {
	zapTracer := NewZapTracer(l.GetZap())
	tracer := &tracelog.TraceLog{
		Logger:   zapTracer,
		LogLevel: tracelog.LogLevelTrace,
	}
	poolConfig.ConnConfig.Tracer = tracer
}
