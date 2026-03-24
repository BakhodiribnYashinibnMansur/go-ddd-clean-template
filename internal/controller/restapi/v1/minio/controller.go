package minio

import (
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"
)

// Controller orchestrates file operations (upload, download, transfer) by
// interacting with the underlying Minio/S3 usecases.
type Controller struct {
	useCase *usecase.UseCase
	logger  logger.Log
}

// New instantiates a new Minio controller with necessary dependencies.
func New(useCase *usecase.UseCase, l logger.Log) *Controller {
	return &Controller{useCase: useCase, logger: l}
}
