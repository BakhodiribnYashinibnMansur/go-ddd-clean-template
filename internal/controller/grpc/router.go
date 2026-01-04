package grpc

import (
	v1 "gct/internal/controller/grpc/v1"
	"gct/internal/usecase"
	"gct/pkg/logger"
	pbgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewRouter -.
func NewRouter(app *pbgrpc.Server, u *usecase.UseCase, l logger.Log) {
	{
		v1.NewUserRoutes(app, u, l)
	}

	reflection.Register(app)
}
