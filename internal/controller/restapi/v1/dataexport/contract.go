package dataexport

import (
	"gct/config"
	ucdataexport "gct/internal/usecase/dataexport"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

type ControllerI interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type Controller struct {
	useCase ucdataexport.UseCaseI
	cfg     *config.Config
	logger  logger.Log
}

func New(uc ucdataexport.UseCaseI, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{useCase: uc, cfg: cfg, logger: l}
}
