package minio

import (
	"gct/internal/usecase"
	"gct/pkg/logger"
)

type Controller struct {
	useCase *usecase.UseCase
	logger  logger.Log
}

func New(useCase *usecase.UseCase, l logger.Log) *Controller {
	return &Controller{useCase: useCase, logger: l}
}
