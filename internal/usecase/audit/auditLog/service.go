package auditlog

import (
	"gct/internal/repo/persistent"
	"gct/internal/shared/infrastructure/logger"
)

type UseCase struct {
	repo   *persistent.Repo
	logger logger.Log
}

func New(r *persistent.Repo, logger logger.Log) UseCaseI {
	return &UseCase{
		repo:   r,
		logger: logger,
	}
}
