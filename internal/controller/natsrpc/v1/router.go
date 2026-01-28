package v1

import (
	"gct/internal/usecase"
	"gct/pkg/broker/nats/natsrpc/server"
	"gct/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// NewUserRoutes -.
func NewUserRoutes(routes map[string]server.CallHandler, u *usecase.UseCase, l logger.Log) {
	_ = &V1{u: u, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	{
		// routes["v1.getUser"] = r.getUser()
	}
}
