package restapi

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/user"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
)

type Controller struct {
	User *user.UserController
}

func NewController(uc *usecase.UseCase, l logger.Log) *Controller {
	return &Controller{
		User: user.NewUserController(uc, l),
	}
}
