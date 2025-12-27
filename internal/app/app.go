// Package app configures and runs application.
package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/evrone/go-clean-template/config"
	amqprpc "github.com/evrone/go-clean-template/internal/controller/amqp_rpc"

	// "github.com/evrone/go-clean-template/internal/controller/grpc"
	natsrpc "github.com/evrone/go-clean-template/internal/controller/nats_rpc"
	"github.com/evrone/go-clean-template/internal/controller/restapi"
	"github.com/evrone/go-clean-template/internal/repo/persistent"
	"github.com/evrone/go-clean-template/internal/repo/webapi"
	"github.com/evrone/go-clean-template/internal/usecase/translation"
	"github.com/evrone/go-clean-template/internal/usecase/user"

	// "github.com/evrone/go-clean-template/pkg/grpcserver"
	"github.com/evrone/go-clean-template/pkg/httpserver"
	"github.com/evrone/go-clean-template/pkg/logger"
	natsRPCServer "github.com/evrone/go-clean-template/pkg/nats/nats_rpc/server"
	"github.com/evrone/go-clean-template/pkg/postgres"
	rmqRPCServer "github.com/evrone/go-clean-template/pkg/rabbitmq/rmq_rpc/server"
	"go.uber.org/zap"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(context.Background(), cfg.App.Environment, cfg.Database.Postgres, l)
	if err != nil {
		l.Fatalw("app - Run - postgres.New", zap.Error(err))
	}
	defer pg.Close()

	// Use-Case
	translationUseCase := translation.New(
		persistent.NewTranslationRepo(pg),
		webapi.New(),
	)

	userUseCase := user.New(
		persistent.NewUserRepo(pg),
	)

	// RabbitMQ RPC Server
	rmqRouter := amqprpc.NewRouter(translationUseCase, l)
	rmqServer, err := rmqRPCServer.New(cfg.Connectivity.RMQ.URL, cfg.Connectivity.RMQ.ServerExchange, rmqRouter, l)
	if err != nil {
		l.Fatalw("app - Run - rmqServer - rmqRPCServer.New", zap.Error(err))
	}

	// NATS RPC Server
	natsRouter := natsrpc.NewRouter(translationUseCase, l)
	natsServer, err := natsRPCServer.New(cfg.Connectivity.NATS.URL, cfg.Connectivity.NATS.ServerExchange, natsRouter, l)
	if err != nil {
		l.Fatalw("app - Run - natsServer - natsRPCServer.New", zap.Error(err))
	}

	// gRPC Server
	// grpcServer := grpcserver.New(l, grpcserver.Port(cfg.Connectivity.GRPC.Port))
	// grpc.NewRouter(grpcServer.App, translationUseCase, l)

	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	restapi.NewRouter(httpServer.App, cfg, translationUseCase, userUseCase, l)

	// Start servers
	rmqServer.Start()
	natsServer.Start()
	// grpcServer.Start()
	httpServer.Start()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Infow("app - Run - signal", zap.String("signal", s.String()))
	case err = <-httpServer.Notify():
		l.Errorw("app - Run - httpServer.Notify", zap.Error(err))
	// case err = <-grpcServer.Notify():
	// 	l.Errorw("app - Run - grpcServer.Notify", zap.Error(err))
	case err = <-rmqServer.Notify():
		l.Errorw("app - Run - rmqServer.Notify", zap.Error(err))
	case err = <-natsServer.Notify():
		l.Errorw("app - Run - natsServer.Notify", zap.Error(err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Errorw("app - Run - httpServer.Shutdown", zap.Error(err))
	}

	// err = grpcServer.Shutdown()
	// if err != nil {
	// 	l.Errorw("app - Run - grpcServer.Shutdown", zap.Error(err))
	// }

	err = rmqServer.Shutdown()
	if err != nil {
		l.Errorw("app - Run - rmqServer.Shutdown", zap.Error(err))
	}

	err = natsServer.Shutdown()
	if err != nil {
		l.Errorw("app - Run - natsServer.Shutdown", zap.Error(err))
	}
}
