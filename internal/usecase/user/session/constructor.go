package session

import (
	"gct/internal/repo/persistent"
	"gct/internal/shared/infrastructure/logger"
)

// UseCase -.
type UseCase struct {
	repo   *persistent.Repo
	logger logger.Log
}

// New -.
func New(r *persistent.Repo, logger logger.Log) UseCaseI {
	return &UseCase{
		repo:   r,
		logger: logger,
	}
}
