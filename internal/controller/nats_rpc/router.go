package v1

import (
	v1 "gct/internal/controller/nats_rpc/v1"
	"gct/internal/usecase"
	"gct/pkg/broker/nats/nats_rpc/server"
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
