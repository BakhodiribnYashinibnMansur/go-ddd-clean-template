package translation

import (
	"gct/config"
	translationUC "gct/internal/usecase/translation"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// ControllerI defines the translation HTTP handler interface.
type ControllerI interface {
	Upsert(ctx *gin.Context)
	Gets(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type Controller struct {
	uc     translationUC.UseCaseI
	cfg    *config.Config
	logger logger.Log
}

func New(uc translationUC.UseCaseI, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{uc: uc, cfg: cfg, logger: l}
}
