package client

import (
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"gct/pkg/db/postgres"
	"gct/pkg/logger"
)

// UserRepo handles user-related database operations.
// SessionRepo handles session-related database operations.
type Repo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
	logger  logger.Log
}

// New creates a new client repository instance.
func New(pg *postgres.Postgres, logger logger.Log) RepoI {
	return &Repo{
		pool:    pg.Pool,
		builder: pg.Builder,
		logger:  logger,
	}
}
