package natsrpc

import (
	v1 "gct/internal/controller/natsrpc/v1"
	"gct/internal/usecase"
	"gct/pkg/broker/nats/natsrpc/server"
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
