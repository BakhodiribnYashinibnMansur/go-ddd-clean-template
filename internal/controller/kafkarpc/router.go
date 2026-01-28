package kafkarpc

import (
	v1 "gct/internal/controller/kafkarpc/v1"
	"gct/internal/usecase"
	"gct/pkg/broker/kafka/kafkarpc/server"
	"gct/pkg/logger"
)

// NewRouter -.
func NewRouter(u *usecase.UseCase, l logger.Log) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)

	{
		v1.NewUserRoutes(routes, u, l)
	}

	return routes
}
