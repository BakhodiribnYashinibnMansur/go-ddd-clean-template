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
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	sharedmw "gct/internal/kernel/infrastructure/middleware"

	// DDD BC middleware
	auditmw "gct/internal/context/iam/supporting/audit/interfaces/http/middleware"
	authzmw "gct/internal/context/iam/generic/authz/interfaces/http/middleware"
	integrationmw "gct/internal/context/admin/supporting/integration/interfaces/http/middleware"
	syserrmw "gct/internal/context/ops/generic/systemerror/interfaces/http/middleware"
	"gct/internal/context/iam/generic/user/application/command"
	usermw "gct/internal/context/iam/generic/user/interfaces/http/middleware"
	userport "gct/internal/context/iam/generic/user/interfaces/port"

	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/asynq"
	"gct/internal/kernel/infrastructure/db/postgres"
	redispkg "gct/internal/kernel/infrastructure/db/redis"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/httpclient"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/metrics"
	"gct/internal/kernel/infrastructure/metrics/latency"
	"gct/internal/kernel/infrastructure/pubsub"
	"gct/internal/kernel/infrastructure/reqlog"
	"gct/internal/kernel/infrastructure/sse"
	jwtpkg "gct/internal/kernel/infrastructure/security/jwt"
	httpserver "gct/internal/kernel/infrastructure/server/http"
	"gct/internal/kernel/infrastructure/tracing"

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
			Level:        cfg.Log.PersistLevel,
			RedisKey:     cfg.Log.RedisKey,
			BatchSize:    cfg.Log.FlushBatchSize,
			Interval:     time.Duration(cfg.Log.FlushInterval) * time.Second,
			RetentionDay: cfg.Log.RetentionDays,
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

	// 2c. External API Log Persistence — same buffer-then-flush pattern, writes
	// to external_api_logs. Consumers obtain apiLogSink via injection to log
	// failed 3rd-party HTTP calls with full request/response context.
	var apiLogSink httpclient.Sink = httpclient.NoopSink{}
	var apiLogRedisSink *httpclient.RedisSink
	var apiLogFlusher *httpclient.Flusher
	if cfg.Log.PersistEnabled && redisclient != nil {
		apiLogRedisSink = httpclient.NewRedisSink(redisclient, httpclient.DefaultRedisKey)
		apiLogSink = apiLogRedisSink
		apiLogFlusher = httpclient.NewFlusher(redisclient, pg.Pool, httpclient.FlusherConfig{
			RedisKey:     httpclient.DefaultRedisKey,
			BatchSize:    cfg.Log.FlushBatchSize,
			Interval:     time.Duration(cfg.Log.FlushInterval) * time.Second,
			RetentionDay: cfg.Log.RetentionDays,
		}, l)
		apiLogFlusher.Start()
		l.Infoc(ctx, "External API log persistence enabled",
			"flush_interval", cfg.Log.FlushInterval,
			"slow_threshold_ms", cfg.Log.APILogSlowThresholdMs,
			"success_sample_rate", cfg.Log.APILogSuccessSampleRate,
		)
	}
	// Thresholds controlling when SUCCESSFUL outgoing calls are persisted
	// alongside failures. Injected into downstream HTTP clients as they are
	// wired up (see telegram.WithAPILogThresholds etc).
	apiLogSlowThreshold := time.Duration(cfg.Log.APILogSlowThresholdMs) * time.Millisecond
	apiLogSuccessSampleRate := cfg.Log.APILogSuccessSampleRate
	_ = apiLogSink               // injected into downstream HTTP clients as they're wired up
	_ = apiLogSlowThreshold      // injected into downstream HTTP clients as they're wired up
	_ = apiLogSuccessSampleRate  // injected into downstream HTTP clients as they're wired up

	// 2d. Incoming HTTP request/response logging — captures every request
	// processed by the Gin engine and persists it to http_request_logs.
	var reqLogSink reqlog.Sink = reqlog.NoopSink{}
	var reqLogRedisSink *reqlog.RedisSink
	var reqLogFlusher *reqlog.Flusher
	if cfg.Log.PersistEnabled && redisclient != nil {
		reqLogRedisSink = reqlog.NewRedisSink(redisclient, reqlog.DefaultRedisKey)
		reqLogSink = reqLogRedisSink
		reqLogFlusher = reqlog.NewFlusher(redisclient, pg.Pool, reqlog.FlusherConfig{
			RedisKey:     reqlog.DefaultRedisKey,
			BatchSize:    cfg.Log.FlushBatchSize,
			Interval:     time.Duration(cfg.Log.FlushInterval) * time.Second,
			RetentionDay: cfg.Log.RetentionDays,
		}, l)
		reqLogFlusher.Start()
		l.Infoc(ctx, "Incoming request log persistence enabled",
			"flush_interval", cfg.Log.FlushInterval,
		)
	}

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
		}, redisclient, l)
		apperrors.SetReporter(alerter)
		alerter.StartPendingLoop(ctx)

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

		// Error correlation chain — groups errors by request_id
		errorChain := apperrors.NewErrorChain(5 * time.Minute)
		errorChain.StartCleanup(ctx)

		// Automated error resolver
		resolver := apperrors.NewResolver()
		// Example: REPO_CONNECTION errors could trigger cache fallback
		// resolver.Register(apperrors.ErrRepoConnection, func(ctx context.Context, err *apperrors.AppError) bool {
		//     // fallback to cache logic here
		//     return true
		// })

		// SLO tracker — 99.9% success rate target, 1 hour window
		sloTracker := apperrors.NewSLOTracker(apperrors.SLOConfig{
			Target: 0.999,
			Window: time.Hour,
			OnBudgetExhausted: func(stats apperrors.SLOStats) {
				alerter.SendError(
					apperrors.New(apperrors.ErrInternal, "").
						WithDetails(fmt.Sprintf("SLO budget exhausted: %.2f%% success rate (target: %.1f%%), %d errors / %d total",
							stats.SuccessRate*100, stats.Target*100, stats.ErrorRequests, stats.TotalRequests)),
				)
			},
		})

		// Register all hooks
		hookMgr := apperrors.GetGlobalHookManager()
		hookMgr.AddHook(func(ctx context.Context, err *apperrors.AppError) {
			apperrors.RateMonitorHook(rateMonitor)(err)
		})
		hookMgr.AddHook(apperrors.ChainHook(errorChain))
		hookMgr.AddHook(apperrors.ResolverHook(resolver))
		hookMgr.AddHook(apperrors.SLOMiddlewareHook(sloTracker))

		_, _, _ = errorChain, resolver, sloTracker // used by hooks above
	}

	// Latency Percentile Tracker
	var latencyTracker *latency.Tracker
	var latencyReporter *latency.Reporter
	if cfg.Metrics.LatencyEnabled {
		p95, _ := time.ParseDuration(cfg.Metrics.LatencyP95Threshold)
		p99, _ := time.ParseDuration(cfg.Metrics.LatencyP99Threshold)
		if p95 == 0 {
			p95 = 200 * time.Millisecond
		}
		if p99 == 0 {
			p99 = 500 * time.Millisecond
		}

		latencyTracker = latency.NewTracker(cfg.Metrics.LatencyWindowSec)

		var alertMgr *latency.AlertManager
		if asynqClient != nil {
			alertMgr = latency.NewAlertManager(
				&asynqClientAdapter{asynqClient},
				latency.AlertConfig{
					P95Threshold: p95,
					P99Threshold: p99,
					Cooldown:     5 * time.Minute,
				})
		}

		interval := time.Duration(cfg.Metrics.LatencyLogIntervalSec) * time.Second
		if interval == 0 {
			interval = 10 * time.Second
		}
		latencyReporter = latency.NewReporter(latencyTracker, alertMgr, interval, cfg.Metrics.LatencyWindowSec, l)
		latencyReporter.Start(ctx)
	}

	// 4. Event Bus — Redis Streams if Redis enabled, otherwise in-memory fallback
	var eventBusInstance application.EventBus
	if redisclient != nil && cfg.SSE.Enabled {
		eventBusInstance = eventbus.NewRedisStreamsEventBus(redisclient, cfg.SSE.StreamMaxLen, l)
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

	// 4.2 Register BC subscribers (event-driven cross-BC coupling).
	// Each BC owns its subscriber wiring; failures are fatal because a
	// missing subscription breaks an event-driven workflow silently.
	if err := dddBCs.Audit.RegisterSubscribers(eventBusInstance); err != nil {
		l.Fatalw("failed to register audit subscribers", "error", err)
	}

	// 4.3 Subscribe session events — Session BC publishes, User BC handles
	subscribeSessionEvents(eventBusInstance, dddBCs.User, l)

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
		bridge := sse.NewBridge(redisclient, sseHub, l)

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
	handler := initRouter(cfg, dddBCs, redisclient, pg, sseHub, metricsProvider, latencyTracker, reqLogSink, l)

	httpServer := httpserver.NewServer()
	startServer(cfg.HTTP.Port, handler, httpServer, l)

	// 8. Wait for shutdown
	waitForSignal(l)
	cancel()

	// 9. Graceful shutdown
	if latencyReporter != nil {
		latencyReporter.Stop()
	}
	if logFlusher != nil {
		l.Infoc(context.Background(), "Flushing remaining logs to database...")
		logFlusher.Stop()
	}
	// Drain in-memory sinks first so any buffered entries reach Redis, then
	// stop the flushers so they pop those entries and COPY FROM to PostgreSQL.
	if apiLogRedisSink != nil {
		apiLogRedisSink.Stop()
	}
	if reqLogRedisSink != nil {
		reqLogRedisSink.Stop()
	}
	if apiLogFlusher != nil {
		l.Infoc(context.Background(), "Flushing remaining external API logs to database...")
		apiLogFlusher.Stop()
	}
	if reqLogFlusher != nil {
		l.Infoc(context.Background(), "Flushing remaining incoming request logs to database...")
		reqLogFlusher.Stop()
	}
	shutdownServer(httpServer, l, cfg.HTTP.ShutdownTimeout)
}

// initRouter configures the Gin engine with DDD middleware and routes.
func initRouter(cfg *config.Config, bcs *DDDBoundedContexts, redisClient *redis.Client, pg *postgres.Postgres, sseHub *sse.Hub, metricsProvider *metrics.Provider, latencyTracker *latency.Tracker, reqLogSink reqlog.Sink, l logger.Log) *gin.Engine {
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

	// === Global shared middleware (with BC-specific middleware injected) ===
	sharedmw.Setup(handler, cfg, redisClient, buildBCMiddleware(bcs, cfg, l), latencyTracker, reqLogSink, l)

	// === Infrastructure routes (swagger, health, static) ===
	var metricsHandler http.Handler
	if metricsProvider != nil {
		metricsHandler = metricsProvider.Handler()
	}
	setupInfraRoutes(handler, cfg, pg.Pool, redisClient, metricsHandler, nil, latencyTracker)

	// === Health check routes ===
	registerHealthRoutes(handler, healthDeps{
		pgPool:    pg.Pool,
		redis:     redisClient,
		asynqAddr: resolveAsynqAddr(cfg),
	})

	// === DDD API routes ===
	authMW := usermw.NewAuthMiddleware(bcs.User.FindSession, bcs.User.FindUserForAuth, cfg, l)
	authUserLookup := userport.NewAuthLookupAdapter(bcs.User.FindUserForAuth)
	authzMiddleware := authzmw.NewAuthzMiddleware(bcs.Authz.CheckAccess, authUserLookup, l)
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

// buildBCMiddleware assembles the bounded-context-specific middleware injected into the shared setup.
func buildBCMiddleware(bcs *DDDBoundedContexts, cfg *config.Config, l logger.Log) *sharedmw.BCMiddleware {
	sysErrMW := syserrmw.NewSystemErrorMiddleware(bcs.SystemError.CreateSystemError, l)
	auditMW := auditmw.NewAuditMiddleware(bcs.Audit.CreateEndpointHistory, bcs.Audit.CreateAuditLog, l)
	sigMW := integrationmw.NewSignatureMiddleware(bcs.Integration.ValidateAPIKey, cfg)
	return &sharedmw.BCMiddleware{
		Recovery:     sysErrMW.Recovery(),
		Persist5xx:   sysErrMW.Persist5xx(),
		AuditHistory: auditMW.EndpointHistory(),
		AuditChange:  auditMW.ChangeAudit(),
		Signature:    sigMW.Validate(),
	}
}

// resolveAsynqAddr returns the Asynq Redis address, falling back to the shared Redis address.
func resolveAsynqAddr(cfg *config.Config) string {
	if !cfg.Asynq.Enabled {
		return ""
	}
	if cfg.Asynq.RedisAddr != "" {
		return cfg.Asynq.RedisAddr
	}
	return cfg.Redis.Addr()
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
	info, err := a.client.EnqueueTask(ctx, taskType, payload)
	if err != nil {
		return nil, fmt.Errorf("app.asynqClientAdapter.EnqueueTask: %w", err)
	}
	return info, nil
}
