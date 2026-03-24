package systemerror

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	tableName = consts.TableSystemError
)

type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repo struct {
	pool    Pool
	builder squirrel.StatementBuilderType
	log     logger.Log
}

func New(pg *postgres.Postgres, log logger.Log) *Repo {
	return &Repo{
		pool:    pg.Pool,
		builder: pg.Builder,
		log:     log,
	}
}
