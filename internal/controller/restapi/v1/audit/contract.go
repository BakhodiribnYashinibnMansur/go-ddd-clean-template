// Package audit encapsulates functionalities for tracking system changes and endpoint access history.
package audit

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/audit/history"
	"gct/internal/controller/restapi/v1/audit/log"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

// Controller acts as a composite handler for various auditing sub-systems.
// It bundles controllers for audit logs and endpoint history into a single interface.
type Controller struct {
	Log     log.ControllerI     // Handles permanent records of business actions.
	History history.ControllerI // Handles transient records of API endpoint access.
}

// New initializes a composite Audit controller by instantiating its dependent sub-controllers.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		Log:     log.New(u, cfg, l),
		History: history.New(u, cfg, l),
	}
}
