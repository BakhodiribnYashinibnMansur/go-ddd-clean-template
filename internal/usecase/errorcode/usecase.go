package errorcode

import (
	"gct/internal/repo"
	errorcoderepo "gct/internal/repo/persistent/postgres/errorcode"
	"gct/pkg/logger"
)

type UseCase struct {
	repo   *errorcoderepo.Repo
	logger logger.Log
}

func New(r *repo.Repo, l logger.Log) UseCaseI {
	return &UseCase{
		repo:   r.Persistent.Postgres.ErrorCode,
		logger: l,
	}
}
