package dashboard

import (
	"context"

	ucdashboard "gct/internal/usecase/dashboard"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"

	"github.com/jackc/pgx/v5"
)

type Pool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repo struct {
	pool   Pool
	logger logger.Log
}

func New(pg *postgres.Postgres, l logger.Log) ucdashboard.Repository {
	return &Repo{pool: pg.Pool, logger: l}
}
