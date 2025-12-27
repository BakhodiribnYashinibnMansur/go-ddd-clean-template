package v1

import (
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/broker/kafka/kafka_rpc/server"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// NewUserRoutes -.
func NewUserRoutes(routes map[string]server.CallHandler, u *usecase.UseCase, l logger.Log) {
	_ = &V1{u: u, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	{
		// routes["v1.getUser"] = r.getUser()
	}
}
