package v1

import (
	v1 "github.com/evrone/go-clean-template/internal/controller/nats_rpc/v1"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/broker/nats/nats_rpc/server"
	"github.com/evrone/go-clean-template/pkg/logger"
)

// NewRouter -.
func NewRouter(u *usecase.UseCase, l logger.Log) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)

	{
		v1.NewUserRoutes(routes, u, l)
	}

	return routes
}
