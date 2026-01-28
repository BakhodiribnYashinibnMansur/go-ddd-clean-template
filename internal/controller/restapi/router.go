// Package restapi centralizes the routing configuration and middleware integration for the HTTP server.
package restapi

import (
	"gct/config"
	"gct/consts"
	"gct/internal/controller/restapi/middleware"
	"gct/internal/controller/restapi/middleware/auth"
	"gct/internal/controller/restapi/v1/admin"
	"gct/internal/controller/restapi/v1/audit"
	"gct/internal/controller/restapi/v1/authz"
	errcode "gct/internal/controller/restapi/v1/errorcode"
	"gct/internal/controller/restapi/v1/featureflag"
	"gct/internal/controller/restapi/v1/minio"
	"gct/internal/controller/restapi/v1/test"
	"gct/internal/controller/restapi/v1/user"
	"gct/internal/usecase"
	webadmin "gct/internal/web/admin"
	"gct/pkg/logger"

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
	// Standard Gin configuration.
	handler.HandleMethodNotAllowed = true

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
	// API V1 & Business Domain Routes
	// ============================================================================
	c := NewController(uc, cfg, l)
	am := auth.NewAuthMiddleware(uc, cfg, l) // Centralized auth handler.

	// Static assets for the Web Administration panel.
	handler.Static("/static", "./internal/web/admin/static")

	// Audit & History: Track API interactions asynchronously.
	auditM := middleware.NewAuditMiddleware(uc, l)
	if cfg.Middleware.AuditHistory {
		handler.Use(auditM.EndpointHistory())
	}
	if cfg.Middleware.AuditChange {
		handler.Use(auditM.ChangeAudit())
	}

	// CSRF: Protection for state-changing requests using HTTP-only cookies.
	csrfM := middleware.HybridMiddleware(l, consts.COOKIE_CSRF_TOKEN)

	// API V1 Group
	h := handler.Group("/api/v1")
	{
		// Business domain routers delegation.
		user.UserRoute(h, c.User, am.AuthClientAccess, csrfM)
		minio.MinioRoute(h, c.Minio, am.AuthClientAccess, csrfM)
		authz.AuthzRoute(h, c.Authz, am.AuthClientAccess, am.AuthClientRefresh, am.Authz, csrfM)
		audit.AuditRoute(h, c.Audit, am.AuthClientAccess)
		errcode.Route(h, c.ErrorCode, am.AuthClientAccess, csrfM)

		// Feature Flag demonstration endpoints.
		featureflag.NewRouter(h, am.AuthClientAccess, l)

		// Administrative system actions (e.g. Linter runner).
		admin.New(l).Register(h, am.AuthAdmin)

		// Test-only endpoints (Environment restricted)
		if cfg.App.Environment != "production" {
			test.TestRoute(h, c.Test)
		}

		// Serve dynamic linter reports.
		handler.Static(lintPath, "./docs/report/linter")

		// Web-based Administrative dashboard.
		webadmin.New(uc, cfg, l).Register(handler.Group("/"), am)
	}
}
