// Package middleware provides shared HTTP middleware for the Gin engine.
// These are generic, cross-cutting concerns with no business logic dependencies.
package middleware

import (
	"net/http"
	"time"

	"gct/config"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/metrics/latency"
	"gct/internal/shared/infrastructure/reqlog"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// BCMiddleware holds optional BC-specific middleware handlers.
// These are injected from bounded contexts and registered alongside shared middleware.
type BCMiddleware struct {
	Recovery     gin.HandlerFunc // SystemError BC: panic recovery with DB persistence
	Persist5xx   gin.HandlerFunc // SystemError BC: 5xx error persistence
	AuditHistory gin.HandlerFunc // Audit BC: endpoint history tracking
	AuditChange  gin.HandlerFunc // Audit BC: state-change audit logging
	Signature    gin.HandlerFunc // Integration BC: request signature validation
}

// Setup registers the standard suite of shared middleware to the Gin engine.
// BC-specific middleware is injected via the bcMW parameter.
func Setup(handler *gin.Engine, cfg *config.Config, redisClient *redis.Client, bcMW *BCMiddleware, latencyTracker *latency.Tracker, reqLogSink reqlog.Sink, l logger.Log) {
	handler.HandleMethodNotAllowed = true

	// 1. Traceability & Logging
	handler.Use(Logger(l))

	// 1.1 Error body capture — logs request body on 4xx/5xx responses
	// so operators can see what the client sent when a request fails.
	// Cheap on the happy path: reads body once but only emits on errors.
	handler.Use(ErrorBody(l))

	// 1.2 Debug-level request body logging (every request, dev only)
	if cfg.Log.IsDebug() {
		handler.Use(DebugBody(l))
	}

	// 2. Security headers
	if cfg.Middleware.Security {
		handler.Use(Security())
	}
	if cfg.Middleware.MetaData {
		handler.Use(FetchMetadata(cfg))
	}

	// 3. OpenTelemetry tracing
	if cfg.Tracing.Enabled {
		handler.Use(otelgin.Middleware(cfg.Tracing.ServiceName))
	}

	// 3.1 OTel HTTP Metrics
	if cfg.Middleware.Metrics && cfg.Metrics.Enabled {
		handler.Use(OTelMetrics(cfg.Tracing.ServiceName))
	}

	// 3.2 Latency Percentile Tracker
	if cfg.Metrics.LatencyEnabled && latencyTracker != nil {
		handler.Use(LatencyTracker(latencyTracker))
	}

	// 4. Resilience (BC-specific: SystemError BC)
	if cfg.Middleware.Recovery && bcMW != nil && bcMW.Recovery != nil {
		handler.Use(bcMW.Recovery)
	}
	if cfg.Middleware.Persist5xx && bcMW != nil && bcMW.Persist5xx != nil {
		handler.Use(bcMW.Persist5xx)
	}

	// 5. CORS
	handler.Use(CORSMiddleware(cfg.CORS))

	// 6. Binding errors → JSON
	handler.Use(BindingErrorMiddleware())

	// 6.1 Strict JSON & Body Limit (2MB)
	binding.EnableDecoderDisallowUnknownFields = true
	handler.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2*1024*1024)
		c.Next()
	})

	// 7. Method Not Allowed
	handler.NoMethod(MethodNotAllowedHandler())

	// 8. Mock (non-production)
	if cfg.Middleware.Mock {
		handler.Use(MockMiddleware(cfg))
	}

	// 8.1 Idempotency
	if redisClient != nil {
		handler.Use(Idempotency(redisClient, l))
	}

	// 9. Rate Limiting
	if cfg.Middleware.RateLimiter && redisClient != nil {
		handler.Use(RateLimiter(cfg.Limiter, redisClient, l))
	}

	// 9.1 Incoming request/response logging — persisted to http_request_logs.
	// Placed AFTER rate-limit so abusive traffic is rejected before we buffer
	// its bodies, and AFTER auth-related middleware once user context is set.
	// Errors and slow requests are always persisted; successful requests are
	// sampled per ReqLogSuccessSampleRate to bound log volume.
	if reqLogSink != nil && cfg.Log.ReqLogEnabled {
		handler.Use(reqlog.Middleware(reqLogSink, reqlog.Config{
			MaxBodyBytes:      cfg.Log.ReqLogMaxBodyBytes,
			SuccessSampleRate: cfg.Log.ReqLogSuccessSampleRate,
			SlowThreshold:     time.Duration(cfg.Log.ReqLogSlowThresholdMs) * time.Millisecond,
			SkipPaths:         cfg.Log.ReqLogSkipPaths,
			SkipPrefixes:      cfg.Log.ReqLogSkipPrefixes,
			BodySuppressPaths: cfg.Log.ReqLogBodySuppressPaths,
		}))
	}

	// 10. Audit (BC-specific: Audit BC)
	if cfg.Middleware.AuditHistory && bcMW != nil && bcMW.AuditHistory != nil {
		handler.Use(bcMW.AuditHistory)
	}
	if cfg.Middleware.AuditChange && bcMW != nil && bcMW.AuditChange != nil {
		handler.Use(bcMW.AuditChange)
	}

	// 11. Signature Verification (BC-specific: Integration BC)
	if cfg.Middleware.Signature && bcMW != nil && bcMW.Signature != nil {
		handler.Use(bcMW.Signature)
	}
}
