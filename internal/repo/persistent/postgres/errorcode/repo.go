package errorcode

import (
	"gct/pkg/db/postgres"
	"gct/pkg/logger"
)

// Repo is the repository for error codes
type Repo struct {
	db     *postgres.Postgres
	logger logger.Log
}

// New creates a new error code repository
func New(pg *postgres.Postgres, l logger.Log) *Repo {
	return &Repo{
		db:     pg,
		logger: l,
	}
}
