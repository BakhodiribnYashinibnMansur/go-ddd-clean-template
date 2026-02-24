package setting

import (
	"gct/internal/usecase/usersetting"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ControllerI interface {
	Gets(c *gin.Context)
	Set(c *gin.Context)
	VerifyPasscode(c *gin.Context)
	RemovePasscode(c *gin.Context)
}

type Controller struct {
	uc usersetting.UseCaseI
	l  logger.Log
}

func New(uc usersetting.UseCaseI, l logger.Log) ControllerI {
	return &Controller{uc: uc, l: l}
}
