package filemetadata

import (
	"context"

	"gct/consts"
	ucfile "gct/internal/usecase/file"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const table = consts.TableFileMetadata

// Pool defines the minimal database interface required by this repo.
type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Repo implements ucfile.Repository against PostgreSQL.
type Repo struct {
	pool    Pool
	builder squirrel.StatementBuilderType
	logger  logger.Log
}

// New creates a new filemetadata Repo.
func New(pg *postgres.Postgres, l logger.Log) ucfile.Repository {
	return &Repo{pool: pg.Pool, builder: pg.Builder, logger: l}
}
