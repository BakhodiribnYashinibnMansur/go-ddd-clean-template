package iprule

import (
	"gct/config"
	uciprule "gct/internal/usecase/iprule"
	"gct/internal/shared/infrastructure/logger"

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
	useCase uciprule.UseCaseI
	cfg     *config.Config
	logger  logger.Log
}

func New(uc uciprule.UseCaseI, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{useCase: uc, cfg: cfg, logger: l}
}
