package restapi

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/audit"
	"gct/internal/controller/restapi/v1/authz"
	"gct/internal/controller/restapi/v1/errorcode"
	"gct/internal/controller/restapi/v1/integration"
	"gct/internal/controller/restapi/v1/minio"
	"gct/internal/controller/restapi/v1/test"
	"gct/internal/controller/restapi/v1/user"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

type Controller struct {
	User        *user.Controller
	Minio       *minio.Controller
	Authz       *authz.Controller
	Audit       *audit.Controller
	ErrorCode   *errorcode.Controller
	Integration integration.ControllerI
	Test        *test.Controller
}

func NewController(uc *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		User:        user.New(uc, cfg, l),
		Minio:       minio.New(uc, l),
		Authz:       authz.New(uc, cfg, l),
		Audit:       audit.New(uc, cfg, l),
		ErrorCode:   errorcode.New(uc.ErrorCode, l),
		Integration: integration.New(uc.Integration, cfg, l),
		Test:        test.New(uc, cfg, l),
	}
}
