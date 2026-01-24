// Package user handles account management and authentication session lifecycle.
package user

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/user/client"
	"gct/internal/controller/restapi/v1/user/session"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

// Controller acts as a composite handler for user-related domains.
// It bundles the client (profile/account) and session (authentication) controllers.
type Controller struct {
	ClientI  client.ControllerI  // Controller for user profile and account actions.
	SessionI session.ControllerI // Controller for active session management.
}

// New initializes a composite User controller by instantiating its specialized sub-handlers.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		ClientI:  client.New(u, cfg, l),
		SessionI: session.New(u, l),
	}
}
