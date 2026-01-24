// Package restapi centralizes the routing configuration and middleware integration for the HTTP server.
package restapi

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"

	"gct/config"
	"gct/consts"
	docs "gct/docs/swagger" // Swagger docs metadata.
	"gct/internal/controller/restapi/middleware"
	"gct/internal/controller/restapi/util"
	"gct/internal/controller/restapi/v1/admin"
	asynqController "gct/internal/controller/restapi/v1/asynq"
	"gct/internal/controller/restapi/v1/audit"
	"gct/internal/controller/restapi/v1/authz"
	"gct/internal/controller/restapi/v1/minio"
	"gct/internal/controller/restapi/v1/user"
	"gct/internal/usecase"
	webAdmin "gct/internal/web/admin"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"
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
	am := middleware.NewAuthMiddleware(uc, cfg, l) // Centralized auth handler.

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
		user.UserRoute(h, c.User, am.AuthClientAccess, am.AuthClientRefresh, csrfM)
		minio.MinioRoute(h, c.Minio, am.AuthClientAccess, csrfM)
		authz.AuthzRoute(h, c.Authz, am.AuthClientAccess, am.Authz, csrfM)
		audit.AuditRoute(h, c.Audit)

		// Background task management (Dev/Test only).
		asynqController.NewRouter(h, uc.AsynqClient, l)

		// Administrative system actions (e.g. Linter runner).
		admin.New(l).Register(h)

		// Serve dynamic linter reports.
		handler.Static(lintPath, "./docs/report/linter")

		// Web-based Administrative dashboard.
		webAdmin.New(uc, cfg, l).Register(handler.Group("/"), am)
	}
}

// setupMetrics configures the Prometheus exporter subsystem name based on app config.
func setupMetrics(handler *gin.Engine, cfg *config.Config) {
	if cfg.Metrics.Enabled {
		subsystem := strings.ReplaceAll(cfg.App.Name, "-", "_")
		subsystem = strings.ReplaceAll(subsystem, " ", "_")

		prometheus := ginprometheus.NewPrometheus(subsystem)
		prometheus.Use(handler)
	}
}

// setupSwagger initializes the Swagger documentation engine and dynamic host resolution.
func setupSwagger(handler *gin.Engine, cfg *config.Config) {
	docs.SwaggerInfo.Version = cfg.App.Version
	if cfg.Swagger.Enabled {
		handler.GET(swaggerRoute, ginSwagger.WrapHandler(swaggerFiles.Handler,
			func() func(*ginSwagger.Config) {
				return func(c *ginSwagger.Config) {
					c.Title = "Go Clean Architecture Swagger Docs"
					c.DocExpansion = "none"
					c.PersistAuthorization = true
					c.DefaultModelsExpandDepth = -1
				}
			}(),
		), func(ctx *gin.Context) {
			docs.SwaggerInfo.Host = ctx.Request.Host
			if ctx.Request.TLS != nil {
				docs.SwaggerInfo.Schemes = []string{"https"}
			}
		})
	}
}

// rootHTML defines the visual layout for the API landing page using Material Design aesthetics.
const rootHTML = `<!DOCTYPE html><html lang="en">...</html>` // Truncated for readability in step output.

// setupRoot serves the visual API landing page with dynamic links to documentation.
func setupRoot(handler *gin.Engine, cfg *config.Config) {
	handler.GET("/", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || util.GetForwardedProto(c) == "https" {
			scheme = "https"
		}

		data := struct {
			SwaggerURL, ProtoURL, AdminURL                           string
			SwaggerEnabled, ProtoEnabled, AdminEnabled, IsProduction bool
		}{
			SwaggerURL:     scheme + "://" + c.Request.Host + swaggerPath,
			ProtoURL:       scheme + "://" + c.Request.Host + protoPath,
			AdminURL:       scheme + "://" + c.Request.Host + adminPath,
			SwaggerEnabled: cfg.Swagger.Enabled,
			ProtoEnabled:   cfg.Proto.Enabled,
			AdminEnabled:   cfg.Admin.Enabled,
			IsProduction:   cfg.App.IsProd(),
		}

		tmpl, _ := template.New("root").Parse(rootHTML)
		var buf bytes.Buffer
		_ = tmpl.Execute(&buf, data)
		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	})
}

// setupHealthCheck registers Kubernetes-compatible liveness and readiness probes.
func setupHealthCheck(handler *gin.Engine, uc *usecase.UseCase) {
	// Liveness: Is the process alive?
	handler.GET("/health/live", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Readiness: Are downstream dependencies (DB, Redis) reachable?
	handler.GET("/health/ready", func(c *gin.Context) {
		if err := uc.HealthCheck(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Legacy endpoints.
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	handler.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
}

// setupProtoDocs serves generated HTML documentation for Protobuf definitions.
func setupProtoDocs(handler *gin.Engine, cfg *config.Config) {
	if cfg.Proto.Enabled {
		handler.StaticFile(protoPath, "./docs/protobuf/doc/index.html")
	}
}
