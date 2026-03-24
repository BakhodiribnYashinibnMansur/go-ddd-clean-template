package relation

import (
	"gct/config"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Controller handles HTTP requests targeting the Relation domain.
type Controller struct {
	u   *usecase.UseCase
	l   logger.Log
	cfg *config.Config
}

// ControllerI defines the standard contract for managing user relations.
type ControllerI interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Gets(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	AddUser(c *gin.Context)
	RemoveUser(c *gin.Context)
}

// New initializes a new Relation controller.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{u: u, l: l, cfg: cfg}
}
