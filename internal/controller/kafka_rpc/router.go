package kafka_rpc

import (
	v1 "github.com/evrone/go-clean-template/internal/controller/kafka_rpc/v1"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/broker/kafka/kafka_rpc/server"
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
