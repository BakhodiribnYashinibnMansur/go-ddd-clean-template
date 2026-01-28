package history

import (
	"gct/internal/repo/schema"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableName = schema.TableEndpointHistory
)

type Repo struct {
	pool    *pgxpool.Pool
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
