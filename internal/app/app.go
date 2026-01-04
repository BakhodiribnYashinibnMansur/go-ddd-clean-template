// Package app configures and runs application.
package app

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"gct/config"
	"gct/internal/controller/restapi"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/pkg/db/postgres"
	redisPkg "gct/pkg/db/redis"
	"gct/pkg/logger"
	httpserver "gct/pkg/server/http"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Context for initialization
	ctx := context.Background()

	// 1. Initialize Postgres
	pg, err := postgres.New(ctx, cfg.App.Environment, cfg.Database.Postgres, l)
	if err != nil {
		l.Fatalw("app - Run - postgres.New", zap.Error(err))
	}
	defer pg.Close()

	// 2. Initialize MinIO
	// minioClient, err := minioPkg.New(cfg.Minio.Endpoint,
	// 	minioPkg.WithCredentials(cfg.Minio.AccessKey, cfg.Minio.SecretKey),
	// 	minioPkg.WithSecure(cfg.Minio.UseSSL),
	// 	minioPkg.WithBucket(cfg.Minio.Bucket, cfg.Minio.Region),
	// )
	// if err != nil {
	// 	l.Fatalw("app - Run - minio.New", zap.Error(err))
	// }
	// 3. Initialize Redis
	redisInstance, err := redisPkg.New(ctx, cfg.App.Environment, cfg.Database.Redis, l)
	if err != nil {
		l.Fatalw("app - Run - redis.New", zap.Error(err))
	}
	defer redisInstance.Close()

	redisClient := redisInstance.Client

	// 4. Initialize Layers
	repositories := repo.New(pg, nil, redisClient, &cfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, cfg)

	// 5. Initialize Router and Server
	handler := initRouter(cfg, useCases, l)
	httpServer := httpserver.NewServer()

	// Start server
	startServer(cfg.HTTP.Port, handler, httpServer, l)

	// 6. Wait for Termination Signal
	waitForSignal(l)

	// 7. Graceful Shutdown
	shutdownServer(httpServer, l)
}

func initRouter(cfg *config.Config, useCases *usecase.UseCase, l logger.Log) *gin.Engine {
	gin.ForceConsoleColor()

	if cfg.App.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	handler := gin.New()
	restapi.NewRouter(handler, cfg, useCases, l)
	return handler
}

func startServer(portStr string, handler *gin.Engine, server *httpserver.Server, l logger.Log) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		l.Fatalw("app - Run - strconv.Atoi", zap.Error(err))
	}

	go func() {
		if err := server.Run(port, handler); err != nil {
			l.Errorw("app - Run - httpServer.Run", zap.Error(err))
		}
	}()
}

func waitForSignal(l logger.Log) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	l.Infow("app - Run - signal received", zap.String("signal", s.String()))
}

func shutdownServer(server *httpserver.Server, l logger.Log) {
	err := server.Shutdown(context.Background())
	if err != nil {
		l.Errorw("app - Run - httpServer.Shutdown", zap.Error(err))
	}
}
