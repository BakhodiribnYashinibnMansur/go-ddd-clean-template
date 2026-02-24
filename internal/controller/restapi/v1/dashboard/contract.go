package dashboard

import (
	ucdashboard "gct/internal/usecase/dashboard"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ControllerI interface {
	Get(c *gin.Context)
}

type Controller struct {
	useCase ucdashboard.UseCaseI
	logger  logger.Log
}

func New(uc ucdashboard.UseCaseI, l logger.Log) ControllerI {
	return &Controller{useCase: uc, logger: l}
}
