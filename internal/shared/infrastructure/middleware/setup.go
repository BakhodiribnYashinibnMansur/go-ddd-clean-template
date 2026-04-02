// Package middleware provides shared HTTP middleware for the Gin engine.
// These are generic, cross-cutting concerns with no business logic dependencies.
package middleware

import (
	"net/http"

	"gct/config"
	"gct/internal/shared/infrastructure/logger"

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
func Setup(handler *gin.Engine, cfg *config.Config, redisClient *redis.Client, bcMW *BCMiddleware, l logger.Log) {
	handler.HandleMethodNotAllowed = true

	// 1. Traceability & Logging
	handler.Use(Logger(l))

	// 1.1 Debug-level request body logging
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
