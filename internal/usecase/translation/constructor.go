package translation

import (
	"gct/internal/shared/infrastructure/logger"
)

type UseCase struct {
	repo   Repository
	logger logger.Log
}

func New(repo Repository, l logger.Log) UseCaseI {
	return &UseCase{
		repo:   repo,
		logger: l,
	}
}
