package v1

import (
	// v1 "github.com/evrone/go-clean-template/docs/proto/v1"

	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
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
