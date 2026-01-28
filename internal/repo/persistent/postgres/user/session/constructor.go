package session

import (
	"context"

	"gct/consts"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const tableName = consts.TableSession

type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// SessionRepo handles session-related database operations.
type Repo struct {
	pool    Pool
	builder squirrel.StatementBuilderType
	logger  logger.Log
}

// New creates a new session repository instance.
func New(pg *postgres.Postgres, logger logger.Log) RepoI {
	return &Repo{
		pool:    pg.Pool,
		builder: pg.Builder,
		logger:  logger,
	}
}
