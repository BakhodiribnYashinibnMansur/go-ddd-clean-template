package system

import (
	"gct/pkg/logger"
)

type Controller struct {
	l logger.Log
}

func New(l logger.Log) *Controller {
	return &Controller{l: l}
}
