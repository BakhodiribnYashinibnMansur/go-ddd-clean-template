// Package restapi centralizes the routing configuration and middleware integration for the HTTP server.
package restapi

import (
	"gct/config"
	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/middleware"
	"gct/internal/controller/restapi/middleware/auth"
	"gct/internal/controller/restapi/v1/admin"
	"gct/internal/controller/restapi/v1/announcement"
	"gct/internal/controller/restapi/v1/audit"
	"gct/internal/controller/restapi/v1/authz"
	"gct/internal/controller/restapi/v1/dashboard"
	"gct/internal/controller/restapi/v1/dataexport"
	"gct/internal/controller/restapi/v1/emailtemplate"
	errcode "gct/internal/controller/restapi/v1/errorcode"
	"gct/internal/controller/restapi/v1/featureflag"
	"gct/internal/controller/restapi/v1/featureflagcrud"
	"gct/internal/controller/restapi/v1/integration"
	"gct/internal/controller/restapi/v1/iprule"
	"gct/internal/controller/restapi/v1/job"
	"gct/internal/controller/restapi/v1/minio"
	"gct/internal/controller/restapi/v1/notification"
	"gct/internal/controller/restapi/v1/ratelimit"
	"gct/internal/controller/restapi/v1/sitesetting"
	"gct/internal/controller/restapi/v1/test"
	"gct/internal/controller/restapi/v1/translation"
	"gct/internal/controller/restapi/v1/user"
	"gct/internal/controller/restapi/v1/webhook"
	"gct/internal/usecase"
	websystem "gct/internal/web/system"
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

	// Silence browser auto-requests (no auth needed).
	handler.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})
	handler.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204)
	})

	// Audit & History: Track API interactions asynchronously.
	auditM := middleware.NewAuditMiddleware(uc, l)
	if cfg.Middleware.AuditHistory {
		handler.Use(auditM.EndpointHistory())
	}
	if cfg.Middleware.AuditChange {
		handler.Use(auditM.ChangeAudit())
	}

	// CSRF: Protection for state-changing requests using HTTP-only cookies.
	csrfM := middleware.HybridMiddleware(l, consts.CookieCsrfToken)

	// API V1 Group
	h := handler.Group("/api/v1")
	{
		// Business domain routers delegation.
		user.UserRoute(h, c.User, am.AuthClientAccess, am.Authz, csrfM)
		minio.MinioRoute(h, c.Minio, am.AuthClientAccess, am.Authz, csrfM)
		authz.AuthzRoute(h, c.Authz, am.AuthClientAccess, am.AuthClientRefresh, am.Authz, csrfM)
		audit.AuditRoute(h, c.Audit, am.AuthClientAccess, am.Authz)
		errcode.Route(h, c.ErrorCode, am.AuthClientAccess, am.Authz, csrfM)
		integration.IntegrationRoute(h, c.Integration, am.AuthClientAccess, am.Authz)
		translation.TranslationRoute(h, c.Translation, am.AuthClientAccess, am.Authz)
		sitesetting.SiteSettingRoute(h, c.SiteSetting, am.AuthClientAccess, am.Authz)

		// Dashboard stats endpoint.
		dashboard.Route(h, c.Dashboard, am.AuthClientAccess, am.Authz)

		// Email Templates CRUD endpoints.
		emailtemplate.Route(h, c.EmailTemplate, am.AuthClientAccess, am.Authz, csrfM)

		// Data Export endpoints.
		dataexport.Route(h, c.DataExport, am.AuthClientAccess, am.Authz, csrfM)

		// Feature Flag CRUD endpoints.
		featureflagcrud.Route(h, c.FeatureFlag, am.AuthClientAccess, am.Authz, csrfM)

		// Rate Limit endpoints.
		ratelimit.Route(h, c.RateLimit, am.AuthClientAccess, am.Authz, csrfM)

		// IP Rules endpoints.
		iprule.Route(h, c.IPRule, am.AuthClientAccess, am.Authz, csrfM)

		// Webhook endpoints.
		webhook.Route(h, c.Webhook, am.AuthClientAccess, am.Authz, csrfM)

		// Job endpoints.
		job.Route(h, c.Job, am.AuthClientAccess, am.Authz, csrfM)

		// Announcement endpoints.
		announcement.Route(h, c.Announcement, am.AuthClientAccess, am.Authz, csrfM)

		// Notification endpoints.
		notification.Route(h, c.Notification, am.AuthClientAccess, am.Authz, csrfM)

		// Feature Flag demonstration endpoints.
		featureflag.NewRouter(h, am.AuthClientAccess, am.Authz, l)

		// Administrative system actions (e.g. Linter runner).
		admin.New(l).Register(h, am.AuthAdmin)

		// Test-only endpoints (Environment restricted)
		if cfg.App.Environment != "production" {
			test.TestRoute(h, c.Test)
		}

		// Serve dynamic linter reports.
		handler.Static(lintPath, "./docs/report/linter")

	}

	// System error reference (JSON API for React admin panel).
	sysCtrl := websystem.New(l)
	h.GET("/system/errors", sysCtrl.GetErrors)
	audit.SystemErrorRoute(h, c.Audit, am.AuthClientAccess, am.Authz, csrfM)

	// Admin panel redirect page — React SPA replaces the Go template admin
	setupAdminRedirect(handler)
}
