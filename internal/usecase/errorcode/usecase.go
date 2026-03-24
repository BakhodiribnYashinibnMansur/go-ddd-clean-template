package errorcode

import (
	"gct/internal/repo"
	"gct/internal/shared/infrastructure/logger"
)

type UseCase struct {
	repo   Repository
	logger logger.Log
}

func New(r *repo.Repo, l logger.Log) UseCaseI {
	return &UseCase{
		repo:   r.Persistent.Postgres.ErrorCode,
		logger: l,
	}
}

// NewWithRepo creates a UseCase with an explicit Repository (useful for testing).
func NewWithRepo(r Repository, l logger.Log) UseCaseI {
	return &UseCase{
		repo:   r,
		logger: l,
	}
}
