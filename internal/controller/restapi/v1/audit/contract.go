package audit

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/audit/history"
	"gct/internal/controller/restapi/v1/audit/log"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

type Controller struct {
	Log     log.ControllerI
	History history.ControllerI
}

func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		Log:     log.New(u, cfg, l),
		History: history.New(u, cfg, l),
	}
}
