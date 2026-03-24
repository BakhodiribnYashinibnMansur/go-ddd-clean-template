package test

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/test/state"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"
)

// Controller acts as a composite handler for test-only operations.
type Controller struct {
	StateI state.ControllerI // Controller for database state management.
}

// New initializes a composite Test controller.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		StateI: state.New(u, cfg, l),
	}
}
