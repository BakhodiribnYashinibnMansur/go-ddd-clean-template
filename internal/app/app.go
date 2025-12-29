// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"gct/config"
	"gct/internal/controller/restapi"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/pkg/db/minio"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"
	httpserver "gct/pkg/server/http"
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
	minioClient, err := minio.New(cfg.Minio.Endpoint,
		minio.WithCredentials(cfg.Minio.AccessKey, cfg.Minio.SecretKey),
		minio.WithSecure(cfg.Minio.UseSSL),
		minio.WithBucket(cfg.Minio.Bucket, cfg.Minio.Region),
	)
	if err != nil {
		l.Fatalw("app - Run - minio.New", zap.Error(err))
	}

	// 3. Initialize Redis
	redisClient := initRedis(ctx, cfg, l)
	defer redisClient.Close()

	// 4. Initialize Layers
	repositories := repo.New(pg, minioClient, redisClient, &cfg.Minio, l)
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

func initRedis(ctx context.Context, cfg *config.Config, l logger.Log) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Database.Redis.Host, cfg.Database.Redis.Port),
		Password: cfg.Database.Redis.Password,
		DB:       0,
	})

	if status := redisClient.Ping(ctx); status.Err() != nil {
		l.Fatalw("app - Run - redis.Ping", zap.Error(status.Err()))
	}
	return redisClient
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
