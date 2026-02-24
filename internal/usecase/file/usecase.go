package file

import (
	"gct/pkg/logger"
)

// UseCase implements UseCaseI.
type UseCase struct {
	repo   Repository
	logger logger.Log
}

// New creates a new file UseCase.
func New(repo Repository, l logger.Log) UseCaseI {
	return &UseCase{repo: repo, logger: l}
}
