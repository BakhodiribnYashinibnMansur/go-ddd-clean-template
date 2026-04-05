// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"fmt"
	"time"

	"gct/config"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultConnAttempts = 10
	defaultConnTimeout  = time.Second
)

// Postgres -.
type Postgres struct {
	Pool    *pgxpool.Pool
	Builder squirrel.StatementBuilderType
}

// New - creates a new Postgres connection.
func New(ctx context.Context, env string, cfg config.Postgres, l logger.Log, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("postgres - New - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.PoolMax)
	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	l.Infoc(ctx, "🔌 Connecting to PostgreSQL...",
		"host", cfg.Host,
		"port", cfg.Port,
		"user", cfg.User,
		"database", cfg.Name,
		"pool_max", cfg.PoolMax,
		"ssl_mode", cfg.SSLMode,
	)

	for _, opt := range opts {
		opt(poolConfig)
	}

	for i := range defaultConnAttempts {
		pg.Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			break
		}
		attemptsLeft := defaultConnAttempts - i - 1
		if attemptsLeft > 0 {
			l.Warnc(ctx, "⚠️  PostgreSQL connection attempt failed, retrying...",
				"attempts_left", attemptsLeft,
				"error", err,
			)
			time.Sleep(defaultConnTimeout)
		}
	}

	if err != nil {
		l.Errorc(ctx, "❌ PostgreSQL connection failed after all attempts",
			"error", err,
			"host", cfg.Host,
			"port", cfg.Port,
			"database", cfg.Name,
			"max_attempts", defaultConnAttempts,
		)
		return nil, fmt.Errorf("postgres - New - pgxpool.NewWithConfig: %w", err)
	}

	l.Infoc(ctx, "✅ PostgreSQL connected successfully",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Name,
		"pool_max", cfg.PoolMax,
		"pool_stats", fmt.Sprintf("total=%d idle=%d", pg.Pool.Stat().TotalConns(), pg.Pool.Stat().IdleConns()),
	)

	return pg, nil
}

// Close - closes the Postgres connection.
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
