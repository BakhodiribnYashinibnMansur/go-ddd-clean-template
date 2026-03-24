package errorcode

import (
	"context"

	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Pool abstracts the database pool for testing.
type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Repo is the repository for error codes
type Repo struct {
	pool    Pool
	builder squirrel.StatementBuilderType
	logger  logger.Log
}

// New creates a new error code repository
func New(pg *postgres.Postgres, l logger.Log) *Repo {
	return &Repo{
		pool:    pg.Pool,
		builder: pg.Builder,
		logger:  l,
	}
}
