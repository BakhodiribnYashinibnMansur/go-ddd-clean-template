// Package app configures and runs application.
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gct/config"
	"gct/consts"
	"gct/internal/controller/restapi"
	"gct/internal/repo"
	"gct/internal/seeder"
	"gct/internal/usecase"
	"gct/internal/usecase/cache"
	"gct/pkg/asynq"
	"gct/pkg/db/postgres"
	redisPkg "gct/pkg/db/redis"
	"gct/pkg/logger"
	httpserver "gct/pkg/server/http"
	"gct/pkg/telemetry"
	"github.com/gin-gonic/gin"
	hibikenAsynq "github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Context for initialization
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Tracing
	shutdown, err := telemetry.InitTracer(ctx, cfg.Tracing)
	if err != nil {
		l.WithContext(ctx).Errorw("failed to init tracer", zap.Error(err))
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			l.WithContext(context.Background()).Errorw("failed to shutdown tracer", zap.Error(err))
		}
	}()

	// 1. Initialize Postgres
	pg, err := postgres.New(ctx, cfg.App.Environment, cfg.Database.Postgres, l)
	if err != nil {
		l.WithContext(ctx).Fatalw("app - Run - postgres.New", zap.Error(err))
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
		l.WithContext(ctx).Fatalw("app - Run - redis.New", zap.Error(err))
	}
	defer redisInstance.Close()

	redisClient := redisInstance.Client

	// 4. Initialize Asynq Client
	asynqClient := asynq.NewClient(
		cfg.Redis.Addr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
		l,
	)
	defer asynqClient.Close()

	// 5. Initialize Layers (pass asynqClient to UseCase)
	repositories := repo.New(pg, nil, redisClient, &cfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, cfg, asynqClient)

	// Initialize Cache service
	cacheService := cache.NewCache(repositories.Persistent.Redis, l)

	// Start listener for postgres notifications
	go pg.Listen(ctx, consts.CacheInvalidationChannel, cacheService.DeletePublicCaches, l)

	// 6. Initialize Asynq Worker (if enabled)
	var asynqWorker *asynq.Worker
	if cfg.Asynq.WorkerEnabled {
		// Set Redis config for Asynq if not explicitly set
		if cfg.Asynq.RedisAddr == "" {
			cfg.Asynq.RedisAddr = cfg.Redis.Addr()
			cfg.Asynq.RedisPassword = cfg.Redis.Password
			cfg.Asynq.RedisDB = cfg.Redis.DB
		}

		asynqWorker = asynq.NewWorker(cfg.Asynq, l)

		// Initialize Seeder for background jobs
		s := seeder.New(repositories, l, cfg)

		// Register task handlers
		handlers := asynq.NewHandlers(l)
		asynqWorker.RegisterHandler(asynq.TypeEmailWelcome, handlers.HandleEmailWelcome)
		asynqWorker.RegisterHandler(asynq.TypeEmailVerification, handlers.HandleEmailVerification)
		asynqWorker.RegisterHandler(asynq.TypeImageResize, handlers.HandleImageResize)
		asynqWorker.RegisterHandler(asynq.TypePushNotification, handlers.HandlePushNotification)

		// Register System Seed handler
		asynqWorker.RegisterHandler(asynq.TypeSystemSeed, func(ctx context.Context, task *hibikenAsynq.Task) error {
			var payload asynq.SeedPayload
			if err := json.Unmarshal(task.Payload(), &payload); err != nil {
				return fmt.Errorf("unmarshal seed payload: %w", err)
			}

			// Prepare custom counts map
			customCounts := make(map[string]int)
			if payload.UsersCount > 0 {
				customCounts["users"] = payload.UsersCount
			}
			if payload.RolesCount > 0 {
				customCounts["roles"] = payload.RolesCount
			}
			if payload.PermissionsCount > 0 {
				customCounts["permissions"] = payload.PermissionsCount
			}
			if payload.PoliciesCount > 0 {
				customCounts["policies"] = payload.PoliciesCount
			}
			if payload.Seed != 0 {
				customCounts["seed"] = int(payload.Seed)
			}
			if payload.ClearData {
				customCounts["clear_data"] = 1
			} else {
				customCounts["clear_data"] = 0
			}

			return s.Seed(ctx, customCounts)
		})

		// Start worker in background
		go func() {
			if err := asynqWorker.Start(); err != nil {
				l.WithContext(ctx).Errorw("failed to start asynq worker", zap.Error(err))
			}
		}()
		defer asynqWorker.Stop()
	}

	// 5. Initialize Router and Server
	handler := initRouter(cfg, useCases, l)
	httpServer := httpserver.NewServer()

	// Start server
	startServer(cfg.HTTP.Port, handler, httpServer, l)

	// 6. Wait for Termination Signal
	waitForSignal(l)
	cancel()

	// 7. Graceful Shutdown
	shutdownServer(httpServer, l, cfg.HTTP.ShutdownTimeout)
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
		l.WithContext(context.Background()).Fatalw("app - Run - strconv.Atoi", zap.Error(err))
	}

	go func() {
		if err := server.Run(port, handler); err != nil {
			l.WithContext(context.Background()).Errorw("app - Run - httpServer.Run", zap.Error(err))
		}
	}()
}

func waitForSignal(l logger.Log) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	l.WithContext(context.Background()).Infow("app - Run - signal received", zap.String("signal", s.String()))
}

func shutdownServer(server *httpserver.Server, l logger.Log, timeout int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		l.WithContext(context.Background()).Errorw("app - Run - httpServer.Shutdown", zap.Error(err))
	} else {
		l.WithContext(context.Background()).Infow("app - Run - httpServer.Shutdown - success")
	}
}
