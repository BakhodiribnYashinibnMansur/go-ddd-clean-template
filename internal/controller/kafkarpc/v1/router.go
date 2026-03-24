package v1

import (
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/broker/kafka/kafkarpc/server"
	"gct/internal/shared/infrastructure/logger"
	"github.com/go-playground/validator/v10"
)

// NewUserRoutes -.
func NewUserRoutes(routes map[string]server.CallHandler, u *usecase.UseCase, l logger.Log) {
	_ = &V1{u: u, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	{
		// routes["v1.getUser"] = r.getUser()
	}
}
