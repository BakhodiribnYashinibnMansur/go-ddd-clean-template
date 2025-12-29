package grpc

import (
	pbgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	v1 "gct/internal/controller/grpc/v1"
	"gct/internal/usecase"
	"gct/pkg/logger"
)

// NewRouter -.
func NewRouter(app *pbgrpc.Server, u *usecase.UseCase, l logger.Log) {
	{
		v1.NewUserRoutes(app, u, l)
	}

	reflection.Register(app)
}
