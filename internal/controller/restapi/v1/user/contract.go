package user

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/user/client"
	"gct/internal/controller/restapi/v1/user/session"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

type Controller struct {
	ClientI  client.ControllerI
	SessionI session.ControllerI
}

func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		ClientI:  client.New(u, cfg, l),
		SessionI: session.New(u, l),
	}
}
