package client

import (
	"github.com/evrone/go-clean-template/internal/repo/persistent"
	"github.com/evrone/go-clean-template/pkg/logger"
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
