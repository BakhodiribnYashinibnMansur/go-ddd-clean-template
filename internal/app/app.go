// Package app contains the highest-level logic for configuring and launching the application.
// It orchestrates the initialization of databases, repositories, usecases, and network servers.
package app

import (
	"context"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gct/config"
	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/internal/usecase/cache"
	"gct/internal/shared/infrastructure/asynq"
	"gct/internal/shared/infrastructure/db/postgres"
	redispkg "gct/internal/shared/infrastructure/db/redis"
	"gct/internal/shared/infrastructure/logger"
	httpserver "gct/internal/shared/infrastructure/server/http"
	"gct/internal/shared/infrastructure/tracing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
	var redisclient *redis.Client
	if cfg.Database.Redis.Enabled {
		redisInstance, err := redispkg.New(ctx, cfg.App.Environment, cfg.Database.Redis, l)
		if err != nil {
			l.Fatalc(ctx, "❌ Failed to initialize Redis", "error", err)
		}
		defer redisInstance.Close()
		redisclient = redisInstance.Client
	} else {
		l.Infoc(ctx, "⚠️ Redis is disabled in configuration")
	}

	// 3. Initialize Asynq Client
	// Used by API handlers to push tasks into background queues.
	var asynqClient *asynq.Client
	if cfg.Asynq.Enabled {
		asynqClient = asynq.NewClient(
			cfg.Redis.Addr(),
			cfg.Redis.Password,
			cfg.Redis.DB,
			l,
		)
		defer asynqClient.Close()
	} else {
		l.Infoc(ctx, "⚠️ Asynq client is disabled in configuration")
	}

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

	// 5.1 Initialize Integration Cache (In-memory)
	if err := useCases.Integration.InitCache(ctx); err != nil {
		l.Errorc(ctx, "⚠️ Failed to initialize integration cache", "error", err)
	}

	// Start background listeners for database-driven cache invalidation events.
	go pg.Listen(ctx, consts.CacheInvalidationChannel, cacheService.DeletePublicCaches, l)
	go pg.Listen(ctx, consts.CacheInvalidationChannel, useCases.Integration.InvalidateCache, l)

	var asynqWorker *asynq.Worker
	if cfg.Asynq.Enabled {
		asynqWorker, err = initAsynqWorker(ctx, cfg, repositories, useCases, l)
		if err != nil {
			l.Errorc(ctx, "⚠️ Failed to initialize Asynq worker", "error", err)
		}
		if asynqWorker != nil {
			defer asynqWorker.Stop()
		}
	} else {
		l.Infoc(ctx, "⚠️ Asynq worker is disabled in configuration")
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
