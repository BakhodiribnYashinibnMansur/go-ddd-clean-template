package policy

import (
	"gct/config"
	"gct/internal/usecase"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Controller handles HTTP requests targeting the Policy domain.
type Controller struct {
	u   *usecase.UseCase
	l   logger.Log
	cfg *config.Config
}

// ControllerI defines the standard contract for managing ABAC policies.
type ControllerI interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Gets(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

// New initializes a new Policy controller.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{u: u, l: l, cfg: cfg}
}
