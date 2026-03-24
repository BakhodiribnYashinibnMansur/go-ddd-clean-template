package emailtemplate

import (
	"context"

	"gct/internal/shared/domain/consts"
	ucemailtemplate "gct/internal/usecase/emailtemplate"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	table = consts.TableEmailTemplates
)

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

func New(pg *postgres.Postgres, l logger.Log) ucemailtemplate.Repository {
	return &Repo{pool: pg.Pool, builder: pg.Builder, logger: l}
}
