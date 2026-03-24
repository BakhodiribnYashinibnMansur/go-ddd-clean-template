package admin

import (
	"gct/internal/shared/infrastructure/logger"
)

// Controller handles administrative tasks for the application.
type Controller struct {
	l logger.Log
}

// New instantiates a new admin controller with the provided logger.
func New(l logger.Log) *Controller {
	return &Controller{
		l: l,
	}
}
