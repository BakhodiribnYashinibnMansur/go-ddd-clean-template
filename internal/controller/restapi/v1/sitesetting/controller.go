package sitesetting

import (
	"gct/config"
	sitesettingUC "gct/internal/usecase/sitesetting"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ControllerI interface {
	Gets(ctx *gin.Context)
	GetByKey(ctx *gin.Context)
	UpdateByKey(ctx *gin.Context)
}

type Controller struct {
	uc     sitesettingUC.UseCaseI
	cfg    *config.Config
	logger logger.Log
}

func New(uc sitesettingUC.UseCaseI, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{uc: uc, cfg: cfg, logger: l}
}
