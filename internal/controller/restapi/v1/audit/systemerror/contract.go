package systemerror

import (
	"gct/internal/usecase/audit/systemerror"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

type ControllerI interface {
	Resolve(c *gin.Context)
}

type Controller struct {
	useCase systemerror.UseCaseI
	logger  logger.Log
}

func New(uc systemerror.UseCaseI, l logger.Log) ControllerI {
	return &Controller{useCase: uc, logger: l}
}
