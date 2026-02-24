// Package user handles account management and authentication session lifecycle.
package user

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/user/client"
	"gct/internal/controller/restapi/v1/user/session"
	"gct/internal/controller/restapi/v1/user/setting"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

// Controller acts as a composite handler for user-related domains.
type Controller struct {
	ClientI  client.ControllerI
	SessionI session.ControllerI
	SettingI setting.ControllerI
}

// New initializes a composite User controller.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		ClientI:  client.New(u, cfg, l),
		SessionI: session.New(u, l),
		SettingI: setting.New(u.UserSetting, l),
	}
}
