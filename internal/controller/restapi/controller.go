package restapi

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/audit"
	"gct/internal/controller/restapi/v1/authz"
	"gct/internal/controller/restapi/v1/minio"
	"gct/internal/controller/restapi/v1/user"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

type Controller struct {
	User  *user.Controller
	Minio *minio.Controller
	Authz *authz.Controller
	Audit *audit.Controller
}

func NewController(uc *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		User:  user.New(uc, cfg, l),
		Minio: minio.New(uc, l),
		Authz: authz.New(uc, cfg, l),
		Audit: audit.New(uc, cfg, l),
	}
}
