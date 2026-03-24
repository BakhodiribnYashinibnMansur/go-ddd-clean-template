package featureflag

import (
	"context"

	"gct/internal/shared/domain/consts"
	ucfeatureflag "gct/internal/usecase/featureflag"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const table = consts.TableFeatureFlags

// Pool defines the database connection pool interface.
type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Repo handles feature flag database operations.
type Repo struct {
	pool    Pool
	builder squirrel.StatementBuilderType
	logger  logger.Log
}

// New creates a new feature flag repository.
func New(pg *postgres.Postgres, l logger.Log) ucfeatureflag.Repository {
	return &Repo{pool: pg.Pool, builder: pg.Builder, logger: l}
}
