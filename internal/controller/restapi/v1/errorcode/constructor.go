package errorcode

import (
	"gct/internal/usecase/errorcode"
	"gct/pkg/logger"
)

type Controller struct {
	useCase errorcode.UseCaseI
	logger  logger.Log
}

func New(useCase errorcode.UseCaseI, logger logger.Log) *Controller {
	return &Controller{
		useCase: useCase,
		logger:  logger,
	}
}
