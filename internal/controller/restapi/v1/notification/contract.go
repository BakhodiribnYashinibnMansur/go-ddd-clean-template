package notification

import (
	"gct/config"
	ucnotification "gct/internal/usecase/notification"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ControllerI interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type Controller struct {
	useCase ucnotification.UseCaseI
	cfg     *config.Config
	logger  logger.Log
}

func New(uc ucnotification.UseCaseI, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{useCase: uc, cfg: cfg, logger: l}
}
