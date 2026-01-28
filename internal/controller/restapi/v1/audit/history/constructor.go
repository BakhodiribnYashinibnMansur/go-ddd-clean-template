package history

import (
	"gct/config"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	u   *usecase.UseCase
	l   logger.Log
	cfg *config.Config
}

type ControllerI interface {
	Gets(c *gin.Context)
}

func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{u: u, l: l, cfg: cfg}
}
