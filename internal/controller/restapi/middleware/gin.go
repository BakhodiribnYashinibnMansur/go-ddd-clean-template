package middleware

import (
	"gct/config"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// GinMiddleware registers the standard suite of foundational middlewares to the Gin engine.
// It applies them based on the configuration flags.
func GinMiddleware(handler *gin.Engine, cfg *config.Config, uc *usecase.UseCase, l logger.Log) {
	// Standard Gin configuration.
	handler.HandleMethodNotAllowed = true

	// Initialize error-tracking middleware.
	sysErrM := NewSystemErrorMiddleware(uc, l)

	// 1. Traceability & Logging: Assign unique IDs and initialize context-aware logger.
	// This helps in tracking a request across different services and logs.
	handler.Use(Logger(l))

	// 2. Security: Apply Helmet headers and Fetch-Metadata protections.
	// These middlewares add security headers to prevent XSS, clickjacking, and other attacks.
	if cfg.Middleware.Security {
		handler.Use(Security())
	}
	// Fetch Metadata middleware helps protect against CSRF and cross-site leaks.
	if cfg.Middleware.MetaData {
		handler.Use(FetchMetadata(cfg))
	}

	// 3. Observability: OpenTelemetry tracing.
	// Enables distributed tracing to monitor application performance and latency.
	if cfg.Tracing.Enabled {
		handler.Use(otelgin.Middleware(cfg.Tracing.ServiceName))
	}

	// 4. Resilience.
	// Recovery middleware recovers from panics, preventing the server from crashing.
	if cfg.Middleware.Recovery {
		handler.Use(sysErrM.Recovery())
	}
	// Persist5xx middleware saves internal server errors (500) to the database for debugging.
	if cfg.Middleware.Persist5xx {
		handler.Use(sysErrM.Persist5xx())
	}

	// 5. Cross-Origin.
	// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) headers.
	// It relies on the CORS configuration provided in the config.yaml.
	handler.Use(CORSMiddleware(cfg.CORS))

	// 6. Binding Error Handler.
	// BindingErrorMiddleware intercepts Gin binding errors and converts them to JSON.
	// This runs early to catch validation errors before they reach Gin's default handler.
	// Works automatically for all endpoints using ShouldBind/ShouldBindJSON.
	handler.Use(BindingErrorMiddleware())

	// 6.1 Strict JSON & Body Limit
	// Enforce strict JSON decoding (no unknown fields) and limit body size to 2MB to prevent DOS.
	binding.EnableDecoderDisallowUnknownFields = true
	handler.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2*1024*1024) // 2MB limit
		c.Next()
	})

	// 7. Method Not Allowed Handler.
	// Handle unsupported HTTP methods with proper 405 response.
	// This runs BEFORE authentication to return 405 instead of 401 for unsupported methods.
	handler.NoMethod(MethodNotAllowedHandler())

	// 8. Maintenance.
	// MockMiddleware allows mocking responses for testing purposes.
	if cfg.Middleware.Mock {
		handler.Use(MockMiddleware(cfg))
	}

	// 8. Idempotency.
	// Handle idempotent requests using Idempotency-Key header.
	// This prevents duplicate side-effects for retried requests.
	if uc.Repo.Persistent.Redis.Client != nil {
		handler.Use(Idempotency(uc.Repo.Persistent.Redis.Client, l))
	}

	// 9. Traffic Control.
	// RateLimiter limits the number of requests a client can make within a time window.
	// It uses Redis to store rate limit counters.
	if cfg.Middleware.RateLimiter && uc.Repo.Persistent.Redis.Client != nil {
		handler.Use(RateLimiter(cfg.Limiter, uc.Repo.Persistent.Redis.Client, l))
	}

	// 10. Audit & History.
	auditM := NewAuditMiddleware(uc, l)
	// EndpointHistory records basic info about every request endpoint hit.
	if cfg.Middleware.AuditHistory {
		handler.Use(auditM.EndpointHistory())
	}
	// ChangeAudit records details about state-changing operations (POST, PUT, DELETE, etc.).
	if cfg.Middleware.AuditChange {
		handler.Use(auditM.ChangeAudit())
	}

	// 11. Signature Verification (Dynamic Keys from DB)
	if cfg.Middleware.Signature {
		handler.Use(SignatureMiddleware(cfg, uc))
	}
}
