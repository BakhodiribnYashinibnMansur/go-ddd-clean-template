package access

import (
	"strings"

	"gct/internal/repo/persistent"
	"gct/pkg/logger"
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

// Helper to check if string slice contains string
func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}
