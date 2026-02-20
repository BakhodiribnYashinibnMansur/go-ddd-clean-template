package integration

import (
	"context"

	"gct/consts"
	"gct/internal/usecase/integration"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	tableIntegrations = consts.TableIntegrations
	tableAPIKeys      = consts.TableAPIKeys
)

// Pool defines the database connection pool interface.
type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Repo handles integration-related database operations.
type Repo struct {
	pool    Pool
	builder squirrel.StatementBuilderType
	logger  logger.Log
}

// New creates a new integration repository instance.
func New(pg *postgres.Postgres, logger logger.Log) integration.Repository {
	return &Repo{
		pool:    pg.Pool,
		builder: pg.Builder,
		logger:  logger,
	}
}
