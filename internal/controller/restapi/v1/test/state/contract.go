package state

import (
	"gct/config"
	"gct/internal/seeder"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Controller handles database state management for tests.
type Controller struct {
	uc     *usecase.UseCase
	cfg    *config.Config
	logger logger.Log
	seeder *seeder.Seeder
}

// ControllerI defines the public contract for test state operations.
type ControllerI interface {
	Reset(c *gin.Context)
	Seed(c *gin.Context)
}

// New initializes a Test State controller.
func New(uc *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{
		uc:     uc,
		cfg:    cfg,
		logger: l,
		seeder: seeder.New(uc.Repo, l, cfg),
	}
}
