package permission

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
	Create(c *gin.Context)
	Get(c *gin.Context)
	Gets(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	AssignScope(c *gin.Context)
	RemoveScope(c *gin.Context)
}

func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{u: u, l: l, cfg: cfg}
}
