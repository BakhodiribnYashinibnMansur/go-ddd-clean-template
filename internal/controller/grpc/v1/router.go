package v1

import (
	"gct/internal/usecase"
	"gct/pkg/logger"
	"github.com/go-playground/validator/v10"
	pbgrpc "google.golang.org/grpc"
)

// NewUserRoutes -.
func NewUserRoutes(app *pbgrpc.Server, u *usecase.UseCase, l logger.Log) {
	_ = &V1{u: u, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	{
		// v1.RegisterUserServer(app, r)
	}
}
