package restapi

import (
	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/system"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/user"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
)

type Controller struct {
	User   *user.UserController
	System *system.Controller
}

func NewController(uc *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		User:   user.NewUserController(uc, cfg, l),
		System: system.NewController(l),
	}
}
