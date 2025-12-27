package client

import (
	"github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/pkg/db/postgres"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
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
