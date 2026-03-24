package systemerror

import (
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"
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
