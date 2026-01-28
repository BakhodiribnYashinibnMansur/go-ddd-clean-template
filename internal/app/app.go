// Package app contains the highest-level logic for configuring and launching the application.
// It orchestrates the initialization of databases, repositories, usecases, and network servers.
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	redispkg "gct/pkg/db/redis"
	"gct/pkg/logger"
	httpserver "gct/pkg/server/http"
	"gct/pkg/telemetry"

	"github.com/gin-gonic/gin"
	hibikenasynq "github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Run initializes the entire application component stack in the correct dependency order.
// This includes telemetry, SQL/NoSQL databases, task queues, and finally the HTTP server.
func Run(cfg *config.Config) {
	// Initialize the centralized logger with the configured severity level.
	l := logger.New(cfg.Log.Level)

	// Log configuration for debugging purposes using zap.Any for clear structure.
	l.Infoc(context.Background(), "🛠️  Application configuration loaded", zap.Any("config", cfg))

	// Context for tracking the initialization phase.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize OpenTelemetry Tracing to monitor application performance and trace requests.
	shutdown, err := telemetry.InitTracer(ctx, cfg.Tracing)
	if err != nil {
		l.Errorc(ctx, "⚠️  Failed to initialize tracer (non-critical)", "error", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			l.Errorc(context.Background(), "⚠️  Failed to shutdown tracer", "error", err)
		}
	}()

	// 1. Initialize PostgresSQL
	// Sets up the connection pool and applies any pre-run logic like migrations checks.
	pg, err := postgres.New(ctx, cfg.App.Environment, cfg.Database.Postgres, l)
	if err != nil {
		l.Fatalc(ctx, "❌ Failed to initialize PostgreSQL", "error", err)
	}
	defer pg.Close()

	// 2. Initialize Redis
	// Provides the primary storage for caching, rate limiting, and session management.
	redisInstance, err := redispkg.New(ctx, cfg.App.Environment, cfg.Database.Redis, l)
	if err != nil {
		l.Fatalc(ctx, "❌ Failed to initialize Redis", "error", err)
	}
	defer redisInstance.Close()

	redisclient := redisInstance.Client

	// 3. Initialize Asynq Client
	// Used by API handlers to push tasks into background queues.
	asynqClient := asynq.NewClient(
		cfg.Redis.Addr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
		l,
	)
	defer asynqClient.Close()

	// 4. Initialize Data Access and Business Layers
	// repositories layer handles raw data retrieval and persistence.
	repositories := repo.New(pg, nil, redisclient, &cfg.Minio, l)
	// useCases layer contains the core business rules and domain logic.
	useCases := usecase.NewUseCase(repositories, l, cfg, asynqClient)

	// 4.1 Initialize Error Codes
	initErrorCodes(ctx, useCases, l)

	// 5. Initialize Reactive Components
	// cacheService manages memory-efficient data invalidation across clusters.
	cacheService := cache.NewCache(repositories.Persistent.Redis, l)

	// Start a background listener for database-driven cache invalidation events.
	go pg.Listen(ctx, consts.CacheInvalidationChannel, cacheService.DeletePublicCaches, l)

	// 6. Initialize Background Workers
	// Asynq workers process time-consuming tasks outside the HTTP request/response cycle.
	var asynqWorker *asynq.Worker
	if cfg.Asynq.WorkerEnabled {
		// Cluster configuration for the worker.
		if cfg.Asynq.RedisAddr == "" {
			cfg.Asynq.RedisAddr = cfg.Redis.Addr()
			cfg.Asynq.RedisPassword = cfg.Redis.Password
			cfg.Asynq.RedisDB = cfg.Redis.DB
		}

		asynqWorker = asynq.NewWorker(cfg.Asynq, l)

		// Setup task handlers (Email, Notifications, Image processing).
		handlers := asynq.NewHandlers(l, useCases.Audit)
		asynqWorker.RegisterHandler(asynq.TypeEmailWelcome, handlers.HandleEmailWelcome)
		asynqWorker.RegisterHandler(asynq.TypeEmailVerification, handlers.HandleEmailVerification)
		asynqWorker.RegisterHandler(asynq.TypeImageResize, handlers.HandleImageResize)
		asynqWorker.RegisterHandler(asynq.TypePushNotification, handlers.HandlePushNotification)
		asynqWorker.RegisterHandler(asynq.TypeAuditLog, handlers.HandleAuditLog)

		// specialized handler for on-demand database seeding.
		s := seeder.New(repositories, l, cfg)
		asynqWorker.RegisterHandler(asynq.TypeSystemSeed, func(ctx context.Context, task *hibikenasynq.Task) error {
			var payload asynq.SeedPayload
			if err := json.Unmarshal(task.Payload(), &payload); err != nil {
				return fmt.Errorf("unmarshal seed payload: %w", err)
			}

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
			customCounts["clear_data"] = 0
			if payload.ClearData {
				customCounts["clear_data"] = 1
			}

			return s.Seed(ctx, customCounts)
		})

		// Launch the worker engine in a managed goroutine.
		go func() {
			if err := asynqWorker.Start(); err != nil {
				l.Errorc(ctx, "❌ Failed to start Asynq worker",
					"error", err,
					"redis_addr", cfg.Asynq.RedisAddr,
					"concurrency", cfg.Asynq.Concurrency,
				)
			}
		}()
		defer asynqWorker.Stop()
	}

	// 7. Initialize Web Router and Persistent Server
	// Translates API requests into usecase calls while applying global middlewares.
	handler := initRouter(cfg, useCases, l)
	httpServer := httpserver.NewServer()

	// Launch the HTTP listener.
	startServer(cfg.HTTP.Port, handler, httpServer, l)

	// 8. Lifecycle Management
	// Blocks execution until an OS termination signal is received.
	waitForSignal(l)
	cancel()

	// 9. Graceful Shutdown
	// Closes network listeners and allows inflight requests to complete within the timeout.
	shutdownServer(httpServer, l, cfg.HTTP.ShutdownTimeout)
}

// initRouter configures the Gin engine with environment-specific modes and routes.
func initRouter(cfg *config.Config, useCases *usecase.UseCase, l logger.Log) *gin.Engine {
	// 1. Separate Gin Mode
	gin.SetMode(cfg.HTTP.GinMode)

	// 2. Control Gin Logs (logga qo'shmaslik imkoniyati)
	if cfg.Log.ShowGin {
		gin.ForceConsoleColor()
		gin.DefaultWriter = logger.NewColorfulWriter(os.Stdout)
		gin.DefaultErrorWriter = logger.NewColorfulWriter(os.Stderr)
	} else {
		gin.DefaultWriter = io.Discard
		gin.DefaultErrorWriter = io.Discard
	}

	handler := gin.New()
	restapi.NewRouter(handler, cfg, useCases, l)
	return handler
}

// startServer converts the port string and launches the listener in a background goroutine.
func startServer(portStr string, handler *gin.Engine, server *httpserver.Server, l logger.Log) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		l.Fatalc(context.Background(), "❌ Invalid HTTP port", "error", err, "port", portStr)
	}

	l.Infoc(context.Background(), "🚀 Starting HTTP server...", "port", port)
	logger.PrintGinBanner(port, gin.Mode())

	go func() {
		if err := server.Run(port, handler); err != nil {
			l.Errorc(context.Background(), "❌ HTTP server error", "error", err, "port", port)
		}
	}()
}

// waitForSignal halts the main thread until SIGINT or SIGTERM is intercepted from the OS.
func waitForSignal(l logger.Log) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	l.Infoc(context.Background(), "✅ Application is running. Press Ctrl+C to stop.")

	s := <-interrupt
	l.Infoc(context.Background(), "🛑 Shutdown signal received", "signal", s.String())
}

// shutdownServer attempts to close the HTTP server and its underlying connections.
func shutdownServer(server *httpserver.Server, l logger.Log, timeout int64) {
	l.Infoc(context.Background(), "⏳ Shutting down HTTP server gracefully...", "timeout_seconds", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		l.Errorc(context.Background(), "❌ HTTP server shutdown failed", "error", err)
	} else {
		l.Infoc(context.Background(), "✅ HTTP server shutdown successfully")
	}
}
