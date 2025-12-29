package v1

import (
	"github.com/go-playground/validator/v10"

	"gct/internal/usecase"
	"gct/pkg/broker/nats/nats_rpc/server"
	"gct/pkg/logger"
)

// NewUserRoutes -.
func NewUserRoutes(routes map[string]server.CallHandler, u *usecase.UseCase, l logger.Log) {
	_ = &V1{u: u, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	{
		// routes["v1.getUser"] = r.getUser()
	}
}
