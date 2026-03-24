package system

import (
	"gct/internal/shared/infrastructure/logger"
)

type Controller struct {
	l logger.Log
}

func New(l logger.Log) *Controller {
	return &Controller{l: l}
}
