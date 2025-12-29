package minio

import (
	"gct/internal/repo"
	"gct/pkg/logger"
)

type UseCase struct {
	repo   *repo.Repo
	logger logger.Log
}

func New(repo *repo.Repo, logger logger.Log) *UseCase {
	return &UseCase{repo: repo, logger: logger}
}
