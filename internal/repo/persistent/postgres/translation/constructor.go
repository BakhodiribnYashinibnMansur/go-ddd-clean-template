package translation

import (
	"context"

	"gct/consts"
	translationUC "gct/internal/usecase/translation"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const tableName = consts.TableTranslations

// Pool defines the minimal database pool interface needed.
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

func New(pg *postgres.Postgres, l logger.Log) translationUC.Repository {
	return &Repo{
		pool:    pg.Pool,
		builder: pg.Builder,
		logger:  l,
	}
}
