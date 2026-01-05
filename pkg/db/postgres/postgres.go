// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"fmt"
	"time"

	"gct/config"
	"gct/pkg/logger"

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
	l.Infof("Postgres Config: Host=%s Port=%d User=%s DB=%s PoolMax=%d", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.PoolMax)
	for _, opt := range opts {
		opt(poolConfig)
	}

	for i := range defaultConnAttempts {
		pg.Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			break
		}
		l.Infof("Postgres is trying to connect, attempts left: %d", defaultConnAttempts-i-1)
		time.Sleep(defaultConnTimeout)
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - New - pgxpool.NewWithConfig: %w", err)
	}

	return pg, nil
}

// Close - closes the Postgres connection.
func (p *Postgres) Close() {
	fmt.Println("DEBUG: Postgres Close called")
	if p.Pool != nil {
		p.Pool.Close()
	}
}
