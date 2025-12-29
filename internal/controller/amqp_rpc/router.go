package v1

import (
	v1 "gct/internal/controller/amqp_rpc/v1"
	"gct/internal/usecase"
	"gct/pkg/broker/rabbitmq/rmq_rpc/server"
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
