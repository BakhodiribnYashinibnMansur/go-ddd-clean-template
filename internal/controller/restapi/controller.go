package restapi

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/minio"
	"gct/internal/controller/restapi/v1/system"
	"gct/internal/controller/restapi/v1/user"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

type Controller struct {
	User   *user.Controller
	System *system.Controller
	Minio  *minio.Controller
}

func NewController(uc *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		User:   user.New(uc, cfg, l),
		System: system.New(l),
		Minio:  minio.New(uc, l),
	}
}
