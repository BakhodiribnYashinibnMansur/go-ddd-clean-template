package permission

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const tableName = consts.TablePermission

type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repo struct {
	pool    Pool
	builder squirrel.StatementBuilderType
	logger  logger.Log
}

func New(pg *postgres.Postgres, logger logger.Log) RepoI {
	return &Repo{
		pool:    pg.Pool,
		builder: pg.Builder,
		logger:  logger,
	}
}
