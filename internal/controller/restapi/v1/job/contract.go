package job

import (
	"gct/config"
	ucjob "gct/internal/usecase/job"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ControllerI interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Trigger(c *gin.Context)
}

type Controller struct {
	useCase ucjob.UseCaseI
	cfg     *config.Config
	logger  logger.Log
}

func New(uc ucjob.UseCaseI, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{useCase: uc, cfg: cfg, logger: l}
}
