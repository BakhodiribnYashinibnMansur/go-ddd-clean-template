// Package restapi implements routing paths. Each services in own file.
package restapi

import (
	"bytes"
	"html/template"
	"net/http"

	"gct/config"
	"gct/consts"
	docs "gct/docs/swagger" // Swagger docs.
	"gct/internal/controller/restapi/middleware"
	"gct/internal/controller/restapi/util"
	"gct/internal/controller/restapi/v1/admin"
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

// NewRouter -.
// Swagger spec:
// @title       Go Clean Template API
// @description Using a translation service as an example
// @version     1.0
// @host        localhost:8080
// @BasePath    /api
func NewRouter(handler *gin.Engine, cfg *config.Config, uc *usecase.UseCase, l logger.Log) {
	// Options
	handler.HandleMethodNotAllowed = true

	// System Error Middleware
	sysErrM := middleware.NewSystemErrorMiddleware(uc, l)

	handler.Use(gin.Logger())
	handler.Use(sysErrM.Recovery())
	handler.Use(sysErrM.Persist5xx())
	handler.Use(middleware.CORSMiddleware())
	handler.Use(middleware.MockMiddleware(cfg))

	// Prometheus metrics
	setupMetrics(handler, cfg)

	// Swagger settings
	setupSwagger(handler, cfg)

	// Proto Docs
	setupProtoDocs(handler, cfg)

	// Root handler
	setupRoot(handler, cfg)

	// K8s probe
	setupHealthCheck(handler)

	// Controller
	c := NewController(uc, cfg, l)

	// Middleware
	am := middleware.NewAuthMiddleware(uc, cfg, l)

	// Audit Middleware
	auditM := middleware.NewAuditMiddleware(uc, l)
	handler.Use(auditM.EndpointHistory())

	// CSRF Middleware
	csrfM := middleware.HybridMiddleware(l, consts.COOKIE_CSRF_TOKEN)

	// Routers
	h := handler.Group("/api/v1")
	{
		user.UserRoute(h, c.User, am.AuthClientAccess, am.AuthClientRefresh, csrfM)
		minio.MinioRoute(h, c.Minio, am.AuthClientAccess, csrfM)
		authz.AuthzRoute(h, c.Authz, am.AuthClientAccess, am.Authz, csrfM)
		audit.AuditRoute(h, c.Audit)

		// Admin API Controller
		adminController := admin.New(l)
		adminController.Register(h)

		// Serve linter reports
		handler.Static("/docs/report/linter", "./docs/report/linter")

		// Web Admin Panel
		adminHandler := webAdmin.New(uc, cfg, l)
		adminHandler.Register(handler.Group("/"), am)
	}
}

func setupMetrics(handler *gin.Engine, cfg *config.Config) {
	if cfg.Metrics.Enabled {
		prometheus := ginprometheus.NewPrometheus("my_service_name")
		prometheus.Use(handler)
	}
}

func setupSwagger(handler *gin.Engine, cfg *config.Config) {
	docs.SwaggerInfo.Version = cfg.App.Version

	if cfg.Swagger.Enabled {
		handler.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			func() func(*ginSwagger.Config) {
				return func(c *ginSwagger.Config) {
					c.Title = "Golang Clean Architecture Swagger Docs"
					c.DocExpansion = "none"
					c.DeepLinking = true
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

const rootHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Clean Template API</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --primary: #6200EE;
            --primary-variant: #3700B3;
            --secondary: #03DAC6;
            --background: #F5F5F5;
            --surface: #FFFFFF;
            --error: #B00020;
            --on-primary: #FFFFFF;
            --on-surface: #000000;
        }
        body {
            font-family: 'Roboto', sans-serif;
            background-color: var(--background);
            margin: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            color: rgba(0, 0, 0, 0.87);
        }
        .container {
            width: 100%;
            max-width: 480px;
            padding: 24px;
        }
        .card {
            background-color: var(--surface);
            border-radius: 8px;
            box-shadow: 0 2px 1px -1px rgba(0,0,0,0.2), 0 1px 1px 0 rgba(0,0,0,0.14), 0 1px 3px 0 rgba(0,0,0,0.12);
            padding: 24px;
            margin-bottom: 24px;
        }
        h1 {
            font-weight: 400;
            font-size: 24px;
            margin: 0 0 16px 0;
            color: var(--primary);
            letter-spacing: 0.18px;
        }
        p {
            font-size: 16px;
            line-height: 1.5;
            color: rgba(0, 0, 0, 0.6);
            margin-bottom: 24px;
        }
        h3 {
            font-weight: 500;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 1.25px;
            color: rgba(0, 0, 0, 0.6);
            margin: 24px 0 16px 0;
        }
        .btn-link {
            display: flex;
            align-items: center;
            padding: 16px;
            border-radius: 4px;
            text-decoration: none;
            color: rgba(0, 0, 0, 0.87);
            background-color: var(--surface);
            transition: background-color 0.2s, box-shadow 0.2s;
            border: 1px solid rgba(0, 0, 0, 0.12);
            margin-bottom: 12px;
        }
        .btn-link:hover {
            background-color: rgba(98, 0, 238, 0.04);
            border-color: var(--primary);
        }
        .btn-link:active {
            background-color: rgba(98, 0, 238, 0.12);
        }
        .icon {
            font-size: 24px;
            margin-right: 16px;
        }
        .text-content {
            display: flex;
            flex-direction: column;
        }
        .title {
            font-weight: 500;
            font-size: 16px;
            letter-spacing: 0.15px;
            color: var(--primary);
        }
        .subtitle {
            font-size: 12px;
            color: rgba(0, 0, 0, 0.6);
            margin-top: 4px;
            word-break: break-all;
        }
        .chip-error {
            background-color: #FDEDED;
            color: #5F2120;
            padding: 12px 16px;
            border-radius: 16px;
            font-size: 14px;
            display: flex;
            align-items: center;
            margin-bottom: 16px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <h1>Go Clean Template API</h1>
            <p>Welcome to the API gateway. Access documentation and administration tools below.</p>

            {{if .IsProduction}}
                <div class="chip-error">
                    <span class="icon" style="font-size: 20px; margin-right: 8px;">⚠️</span>
                    <div>
                        <strong>Production Environment</strong><br>
                        Documentation is disabled for security.
                    </div>
                </div>
            {{else if or .SwaggerEnabled .ProtoEnabled}}
                <h3>Documentation</h3>
                
                {{if .SwaggerEnabled}}
                <a href="{{.SwaggerURL}}" class="btn-link">
                    <span class="icon">📄</span>
                    <div class="text-content">
                        <span class="title">Swagger UI</span>
                        <span class="subtitle">Interactive REST API Documentation</span>
                    </div>
                </a>
                {{end}}
                
                {{if .ProtoEnabled}}
                <a href="{{.ProtoURL}}" class="btn-link">
                    <span class="icon">🛠️</span>
                    <div class="text-content">
                        <span class="title">Protobuf Docs</span>
                        <span class="subtitle">gRPC Protocol Buffer Definitions</span>
                    </div>
                </a>
                {{end}}
            {{end}}

            {{if .AdminEnabled}}
            <h3>Administration</h3>
            <a href="{{.AdminURL}}" class="btn-link">
                <span class="icon">⚙️</span>
                <div class="text-content">
                    <span class="title">Web Admin Panel</span>
                    <span class="subtitle">Manage Users, Policies & System</span>
                </div>
            </a>
            {{end}}

            {{if and (not .SwaggerEnabled) (not .ProtoEnabled) (not .AdminEnabled)}}
                <div class="chip-error">
                    <span class="icon" style="font-size: 20px; margin-right: 8px;">❌</span>
                    No services enabled. Check configuration.
                </div>
            {{end}}
        </div>
    </div>
</body>
</html>
`

func setupRoot(handler *gin.Engine, cfg *config.Config) {
	handler.GET("/", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || util.GetForwardedProto(c) == "https" {
			scheme = "https"
		}
		host := c.Request.Host

		data := struct {
			SwaggerURL     string
			ProtoURL       string
			AdminURL       string
			SwaggerEnabled bool
			ProtoEnabled   bool
			AdminEnabled   bool
			IsProduction   bool
		}{
			SwaggerURL:     scheme + "://" + host + "/swagger/index.html",
			ProtoURL:       scheme + "://" + host + "/docs/proto",
			AdminURL:       scheme + "://" + host + "/admin/dashboard",
			SwaggerEnabled: cfg.Swagger.Enabled,
			ProtoEnabled:   cfg.Proto.Enabled,
			AdminEnabled:   cfg.Admin.Enabled,
			IsProduction:   cfg.App.IsProd(),
		}

		tmpl, err := template.New("root").Parse(rootHTML)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	})
}

func setupHealthCheck(handler *gin.Engine) {
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	handler.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
}

func setupProtoDocs(handler *gin.Engine, cfg *config.Config) {
	if cfg.Proto.Enabled {
		handler.StaticFile("/docs/proto", "./docs/protobuf/doc/index.html")
	}
}
