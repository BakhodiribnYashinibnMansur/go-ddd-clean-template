package user

import (
	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/user/client"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/user/session"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
)

type UserController struct {
	ClientI  client.ControllerI
	SessionI session.ControllerI
}

func NewUserController(u *usecase.UseCase, cfg *config.Config, l logger.Log) *UserController {
	return &UserController{
		ClientI:  client.New(u, cfg, l),
		SessionI: session.New(u, l),
	}
}
