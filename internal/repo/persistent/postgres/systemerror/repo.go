package systemerror

import (
	"gct/pkg/db/postgres"
	"gct/pkg/logger"
)

// Repo handles system error logging operations
type Repo struct {
	db     *postgres.Postgres
	logger logger.Log
}

// New creates a new system error repository instance
func New(db *postgres.Postgres, logger logger.Log) *Repo {
	return &Repo{
		db:     db,
		logger: logger,
	}
}
