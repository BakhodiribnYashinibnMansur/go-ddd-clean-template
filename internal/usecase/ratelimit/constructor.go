package ratelimit

import (
	"gct/config"
	"gct/internal/shared/infrastructure/logger"
)

type UseCase struct {
	repo   Repository
	logger logger.Log
	cfg    *config.Config
}

func New(repo Repository, l logger.Log, cfg *config.Config) UseCaseI {
	return &UseCase{repo: repo, logger: l, cfg: cfg}
}
