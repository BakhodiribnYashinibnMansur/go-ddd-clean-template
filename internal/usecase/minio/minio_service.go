package minio

import (
	"gct/internal/repo"
	"gct/internal/shared/infrastructure/logger"
)

type UseCase struct {
	repo   *repo.Repo
	logger logger.Log
}

func New(repo *repo.Repo, logger logger.Log) Interface {
	return &UseCase{repo: repo, logger: logger}
}
