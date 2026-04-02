// Package app contains the highest-level logic for configuring and launching the application.
// It orchestrates the initialization of databases, DDD bounded contexts, and the HTTP server.
package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gct/config"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	sharedmw "gct/internal/shared/infrastructure/middleware"

	// DDD BC middleware
	auditmw "gct/internal/audit/interfaces/http/middleware"
	authzmw "gct/internal/authz/interfaces/http/middleware"
	integrationmw "gct/internal/integration/interfaces/http/middleware"
	syserrmw "gct/internal/systemerror/interfaces/http/middleware"
	"gct/internal/user/application/command"
	usermw "gct/internal/user/interfaces/http/middleware"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/asynq"
	"gct/internal/shared/infrastructure/db/postgres"
	redispkg "gct/internal/shared/infrastructure/db/redis"
	"gct/internal/shared/infrastructure/eventbus"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/metrics"
	"gct/internal/shared/infrastructure/pubsub"
	"gct/internal/shared/infrastructure/sse"
	jwtpkg "gct/internal/shared/infrastructure/security/jwt"
	httpserver "gct/internal/shared/infrastructure/server/http"
	"gct/internal/shared/infrastructure/tracing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Run initializes the entire application component stack in the correct dependency order.
func Run(cfg *config.Config) {
	l := logger.NewWithFormat(cfg.Log.Level, cfg.Log.Format)
	if cfg.Log.SlowOpThresholdMs > 0 {
		logger.SetSlowOpThreshold(time.Duration(cfg.Log.SlowOpThresholdMs) * time.Millisecond)
	}
	l.Infoc(context.Background(), "Application configuration loaded", zap.Any("config", cfg))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// OpenTelemetry Tracing
	shutdown, err := telemetry.InitTracer(ctx, cfg.Tracing)
	if err != nil {
		l.Errorc(ctx, "Failed to initialize tracer (non-critical)", "error", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			l.Errorc(context.Background(), "Failed to shutdown tracer", "error", err)
		}
	}()

	// 1. OTel Metrics
	var metricsProvider *metrics.Provider
	if cfg.Metrics.Enabled {
		metricsProvider, err = metrics.NewProvider(cfg.Tracing.ServiceName)
		if err != nil {
			l.Errorc(ctx, "Failed to initialize metrics provider (non-critical)", "error", err)
		} else {
			defer metricsProvider.Shutdown(context.Background())
		}
	}

	// 2. PostgreSQL
	pgOpts := []postgres.Option{}
	if cfg.Metrics.Enabled {
		slowThreshold, parseErr := time.ParseDuration(cfg.Metrics.SlowQueryThreshold)
		if parseErr != nil {
			slowThreshold = 100 * time.Millisecond
		}
		pgOpts = append(pgOpts, postgres.WithMetricsTracer(l, slowThreshold))
	}
	pg, err := postgres.New(ctx, cfg.App.Environment, cfg.Database.Postgres, l, pgOpts...)
	if err != nil {
		l.Fatalc(ctx, "Failed to initialize PostgreSQL", "error", err)
	}
	defer pg.Close()

	// Register DB pool metrics
	if cfg.Metrics.Enabled {
		if poolErr := metrics.RegisterPoolMetrics(pg.Pool, cfg.Tracing.ServiceName); poolErr != nil {
			l.Errorc(ctx, "Failed to register DB pool metrics (non-critical)", "error", poolErr)
		}
	}

	// 2. Redis
	var redisclient *redis.Client
	if cfg.Database.Redis.Enabled {
		redisInstance, err := redispkg.New(ctx, cfg.App.Environment, cfg.Database.Redis, l)
		if err != nil {
			l.Fatalc(ctx, "Failed to initialize Redis", "error", err)
		}
		defer redisInstance.Close()
		redisclient = redisInstance.Client
	} else {
		l.Infoc(ctx, "Redis is disabled in configuration")
	}

	// 2b. Log Persistence — buffer in Redis, flush to PostgreSQL via COPY FROM
	var logFlusher *logger.Flusher
	if cfg.Log.PersistEnabled && redisclient != nil {
		persistCfg := logger.PersistConfig{
			Level:     cfg.Log.PersistLevel,
			RedisKey:  cfg.Log.RedisKey,
			BatchSize: cfg.Log.FlushBatchSize,
			Interval:  time.Duration(cfg.Log.FlushInterval) * time.Second,
		}
		redisSink := logger.NewRedisSink(redisclient, persistCfg)
		l = logger.WithPersistCore(l, redisSink)

		logFlusher = logger.NewFlusher(redisclient, pg.Pool, persistCfg, l)
		logFlusher.Start()
		l.Infoc(ctx, "Log persistence enabled",
			"level", cfg.Log.PersistLevel,
			"flush_interval", cfg.Log.FlushInterval,
		)
	}
	_ = logFlusher // used in shutdown below

	// 3. Asynq Client
	var asynqClient *asynq.Client
	if cfg.Asynq.Enabled {
		asynqClient = asynq.NewClient(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB, l)
		defer asynqClient.Close()
	} else {
		l.Infoc(ctx, "Asynq client is disabled in configuration")
	}
	_ = asynqClient // used by asynq worker below

	// Error alerter — sends CRITICAL/HIGH errors to Telegram via Asynq
	if asynqClient != nil {
		alerter := apperrors.NewAlerter(&asynqClientAdapter{asynqClient}, apperrors.AlerterConfig{
			MinSeverity:    apperrors.SeverityHigh,
			DebouncePeriod: time.Minute,
		})
		apperrors.SetReporter(alerter)

		// Error rate monitor — alerts when error rate exceeds threshold
		rateMonitor := apperrors.NewRateMonitor(apperrors.RateMonitorConfig{
			Window:    time.Minute,
			Threshold: 10,
			OnBreach: func(code string, count int) {
				alerter.SendError(
					apperrors.New(apperrors.ErrInternal, "").
						WithDetails(fmt.Sprintf("Error rate breach: %s occurred %d times in 1 minute", code, count)),
				)
			},
		})

		hookMgr := apperrors.GetGlobalHookManager()
		hookMgr.AddHook(func(ctx context.Context, err *apperrors.AppError) {
			apperrors.RateMonitorHook(rateMonitor)(err)
		})
	}

	// 4. Event Bus — Redis Streams if Redis enabled, otherwise in-memory fallback
	var eventBusInstance application.EventBus
	if redisclient != nil && cfg.SSE.Enabled {
		eventBusInstance = eventbus.NewRedisStreamsEventBus(redisclient, cfg.SSE.StreamMaxLen)
		l.Infoc(ctx, "EventBus: Redis Streams")
	} else {
		eventBusInstance = eventbus.NewInMemoryEventBus()
		l.Infoc(ctx, "EventBus: In-Memory (dev mode)")
	}

	jwtPrivateKey, err := jwtpkg.ParseRSAPrivateKey(cfg.JWT.PrivateKey)
	if err != nil {
		l.Fatalw("failed to parse RSA private key for DDD", "error", err)
	}

	// Business Metrics
	var businessMetrics *metrics.BusinessMetrics
	if cfg.Metrics.Enabled {
		businessMetrics = metrics.NewBusinessMetrics(cfg.Tracing.ServiceName)
	}

	dddBCs, err := NewDDDBoundedContexts(ctx, pg.Pool, eventBusInstance, l, businessMetrics, command.JWTConfig{
		PrivateKey: jwtPrivateKey,
		Issuer:     cfg.JWT.Issuer,
		AccessTTL:  cfg.JWT.AccessTTL,
		RefreshTTL: cfg.JWT.RefreshTTL,
	})
	if err != nil {
		l.Fatalw("failed to initialize DDD bounded contexts", "error", err)
	}

	// 4.1 Initialize Error Codes
	initErrorCodes(ctx, dddBCs.ErrorCode, eventBusInstance, l)

	// 5. Integration Cache
	if err := dddBCs.Integration.Cache.InitCache(ctx); err != nil {
		l.Errorc(ctx, "Failed to initialize integration cache", "error", err)
	}
	go pg.Listen(ctx, consts.CacheInvalidationChannel, dddBCs.Integration.Cache.InvalidateCache, l)

	// 6. Asynq Worker
	if cfg.Asynq.Enabled {
		asynqWorker, err := initAsynqWorker(ctx, cfg, pg.Pool, dddBCs.Audit.CreateAuditLog, l, nil, nil)
		if err != nil {
			l.Errorc(ctx, "Failed to initialize Asynq worker", "error", err)
		}
		if asynqWorker != nil {
			defer asynqWorker.Stop()
		}
	} else {
		l.Infoc(ctx, "Asynq worker is disabled in configuration")
	}

	// 7. SSE Hub + Bridge + Pub/Sub Listeners
	var sseHub *sse.Hub
	if redisclient != nil && cfg.SSE.Enabled {
		sseHub = sse.NewHub(cfg.SSE.ClientBufferSize)
		bridge := sse.NewBridge(redisclient, sseHub)

		// SSE bridges for stream → hub
		go bridge.Listen(ctx, "audit", "signal:audit_log.created", "stream:audit_log.created")
		go bridge.Listen(ctx, "monitoring", "signal:system_error.recorded", "stream:system_error.recorded")

		// Notification bridge (user-specific routing)
		go func() {
			ps := redisclient.Subscribe(ctx, "signal:notification.sent")
			defer ps.Close()

			lastID := "0"
			psCh := ps.Channel()
			for {
				select {
				case <-ctx.Done():
					return
				case _, ok := <-psCh:
					if !ok {
						return
					}
					msgs, err := redisclient.XRead(ctx, &redis.XReadArgs{
						Streams: []string{"stream:notification.sent", lastID},
						Count:   100,
					}).Result()
					if err != nil {
						continue
					}
					for _, stream := range msgs {
						for _, msg := range stream.Messages {
							lastID = msg.ID
							data, ok := msg.Values["data"]
							if !ok {
								continue
							}
							// Broadcast to all connected notification clients
							sseHub.Broadcast("notifications", sse.Message{
								ID:    msg.ID,
								Event: "notification",
								Data:  []byte(data.(string)),
							})
						}
					}
				}
			}
		}()

		// Internal listeners
		ffListener := pubsub.NewFeatureFlagListener(redisclient, func() {
			dddBCs.FeatureFlag.Evaluator.Invalidate(context.Background())
			l.Info("Feature flag cache invalidated via Pub/Sub")
		})
		go ffListener.Start(ctx)

		cacheListener := pubsub.NewCacheInvalidationListener(redisclient, func(key string) {
			l.Infoc(ctx, "Cache invalidated via Pub/Sub", "key", key)
		})
		go cacheListener.Start(ctx)

		l.Infoc(ctx, "SSE Hub and Pub/Sub listeners started")
	}

	// 8. HTTP Router (pure DDD)
	handler := initRouter(cfg, dddBCs, redisclient, pg, sseHub, metricsProvider, l)

	httpServer := httpserver.NewServer()
	startServer(cfg.HTTP.Port, handler, httpServer, l)

	// 8. Wait for shutdown
	waitForSignal(l)
	cancel()

	// 9. Graceful shutdown
	if logFlusher != nil {
		l.Infoc(context.Background(), "Flushing remaining logs to database...")
		logFlusher.Stop()
	}
	shutdownServer(httpServer, l, cfg.HTTP.ShutdownTimeout)
}

// initRouter configures the Gin engine with DDD middleware and routes.
func initRouter(cfg *config.Config, bcs *DDDBoundedContexts, redisClient *redis.Client, pg *postgres.Postgres, sseHub *sse.Hub, metricsProvider *metrics.Provider, l logger.Log) *gin.Engine {
	gin.SetMode(cfg.HTTP.GinMode)

	if cfg.Log.ShowGin {
		gin.ForceConsoleColor()
		gin.DefaultWriter = logger.NewColorfulWriter(os.Stdout)
		gin.DefaultErrorWriter = logger.NewColorfulWriter(os.Stderr)
	} else {
		gin.DefaultWriter = io.Discard
		gin.DefaultErrorWriter = io.Discard
	}

	handler := gin.New()

	// === BC-specific middleware (injected into shared setup) ===
	sysErrMW := syserrmw.NewSystemErrorMiddleware(bcs.SystemError.CreateSystemError, l)
	auditMW := auditmw.NewAuditMiddleware(bcs.Audit.CreateEndpointHistory, bcs.Audit.CreateAuditLog, l)
	sigMW := integrationmw.NewSignatureMiddleware(bcs.Integration.ValidateAPIKey, cfg)

	bcMW := &sharedmw.BCMiddleware{
		Recovery:     sysErrMW.Recovery(),
		Persist5xx:   sysErrMW.Persist5xx(),
		AuditHistory: auditMW.EndpointHistory(),
		AuditChange:  auditMW.ChangeAudit(),
		Signature:    sigMW.Validate(),
	}

	// === Global shared middleware ===
	sharedmw.Setup(handler, cfg, redisClient, bcMW, l)

	// === Infrastructure routes (swagger, health, static) ===
	var metricsHandler http.Handler
	if metricsProvider != nil {
		metricsHandler = metricsProvider.Handler()
	}
	setupInfraRoutes(handler, cfg, pg.Pool, redisClient, metricsHandler, nil)

	// === Health check routes ===
	registerHealthRoutes(handler, healthDeps{
		pgPool: pg.Pool,
		redis:  redisClient,
	})

	// === DDD API routes ===
	authMW := usermw.NewAuthMiddleware(bcs.User.FindSession, bcs.User.FindUserForAuth, cfg, l)
	authzMiddleware := authzmw.NewAuthzMiddleware(bcs.Authz.CheckAccess, bcs.User.FindUserForAuth, l)
	csrfMW := sharedmw.HybridMiddleware(l, consts.CookieCsrfToken)

	// Error dashboard (admin API)
	registerErrorDashboardRoutes(handler.Group("/api/v1"))

	RegisterDDDRoutes(handler, bcs, authMW.AuthClientAccess, authzMiddleware.Authz, csrfMW, l)

	// === SSE streaming routes ===
	if sseHub != nil {
		heartbeat := time.Duration(cfg.SSE.HeartbeatInterval) * time.Second
		sseHandler := sse.NewHandler(sseHub, heartbeat)
		sse.RegisterRoutes(handler, sseHandler, authMW.AuthClientAccess, authzMiddleware.Authz)
	}

	return handler
}

// startServer converts the port string and launches the listener in a background goroutine.
func startServer(portStr string, handler *gin.Engine, server *httpserver.Server, l logger.Log) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		l.Fatalc(context.Background(), "Invalid HTTP port", "error", err, "port", portStr)
	}

	l.Infoc(context.Background(), "Starting HTTP server...", "port", port)
	logger.PrintGinBanner(port, gin.Mode())

	go func() {
		if err := server.Run(port, handler); err != nil {
			l.Errorc(context.Background(), "HTTP server error", "error", err, "port", port)
		}
	}()
}

// waitForSignal halts the main thread until SIGINT or SIGTERM is intercepted from the OS.
func waitForSignal(l logger.Log) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	l.Infoc(context.Background(), "Application is running. Press Ctrl+C to stop.")
	s := <-interrupt
	l.Infoc(context.Background(), "Shutdown signal received", "signal", s.String())
}

// shutdownServer attempts to close the HTTP server and its underlying connections.
func shutdownServer(server *httpserver.Server, l logger.Log, timeout int64) {
	l.Infoc(context.Background(), "Shutting down HTTP server gracefully...", "timeout_seconds", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		l.Errorc(context.Background(), "HTTP server shutdown failed", "error", err)
	} else {
		l.Infoc(context.Background(), "HTTP server shutdown successfully")
	}
}

// asynqClientAdapter adapts *asynq.Client to the apperrors.TaskEnqueuer interface.
// apperrors.TaskEnqueuer uses opts ...any for testability, while the real asynq client
// uses opts ...asynq.Option — this adapter bridges the two.
type asynqClientAdapter struct {
	client *asynq.Client
}

func (a *asynqClientAdapter) EnqueueTask(ctx context.Context, taskType string, payload any, opts ...any) (any, error) {
	return a.client.EnqueueTask(ctx, taskType, payload)
}
