// Package restapi centralizes the routing configuration and middleware integration for the HTTP server.
package restapi

import (
	"gct/config"
	"gct/internal/controller/restapi/middleware"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Named constants for various internal documentation and administrative paths.
const (
	swaggerPath  = "/docs/swagger/index.html"
	swaggerRoute = "/docs/swagger/*any"
	protoPath    = "/docs/proto"
	adminPath    = "/admin/dashboard"
	lintPath     = "/docs/linter"
)

// NewRouter constructs the entire Gin routing table, applying global middlewares
// and registering service-specific controllers.
func NewRouter(handler *gin.Engine, cfg *config.Config, uc *usecase.UseCase, l logger.Log) {
	// ============================================================================
	// Global Middleware Stack (Order of execution matters)
	// ============================================================================

	// Centralized middleware registration based on config.
	middleware.GinMiddleware(handler, cfg, uc, l)

	// ============================================================================
	// Infrastructure Services
	// ============================================================================
	if cfg.Middleware.Metrics {
		setupMetrics(handler, cfg) // Prometheus endpoint.
	}
	setupSwagger(handler, cfg)   // Swagger UI.
	setupProtoDocs(handler, cfg) // Protobuf documentation.
	setupRoot(handler, cfg)      // Greeting page.

	if cfg.Middleware.HealthCheck {
		setupHealthCheck(handler, uc) // K8s Liveness/Readiness probes.
	}

	// ============================================================================
	// Browser & Static Routes
	// ============================================================================

	// Silence browser auto-requests (no auth needed).
	handler.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})
	handler.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	// Serve dynamic linter reports.
	handler.Static(lintPath, "./docs/report/linter")

	// Admin panel redirect page
	setupAdminRedirect(handler)
}
