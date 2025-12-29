// Package restapi implements routing paths. Each services in own file.
package restapi

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"gct/config"
	docs "gct/docs/swagger" // Swagger docs.
	"gct/internal/controller/restapi/middleware"
	"gct/internal/controller/restapi/v1/minio"
	"gct/internal/controller/restapi/v1/user"
	"gct/internal/usecase"
	"gct/pkg/logger"
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

	middleware.GinMiddleware(handler)

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

	// Routers
	h := handler.Group("/api")
	{
		user.UserRoute(h, c.User, am.AuthClientAccess)
		minio.MinioRoute(h, c.Minio, am.AuthClientAccess)
		h.GET("/system/errors", c.System.GetErrors)
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
<html>
<head>
	<title>Go Clean Template API</title>
	<style>
		body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; padding: 40px; line-height: 1.6; background-color: #f8f9fa; color: #333; }
		.container { max-width: 800px; margin: 0 auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
		h1 { color: #2c3e50; border-bottom: 2px solid #eee; padding-bottom: 20px; font-weight: 300; }
		p { font-size: 1.1em; color: #666; margin-bottom: 30px; }
		.link-group { margin-bottom: 30px; }
		.link-item { display: block; padding: 15px; border: 1px solid #e1e4e8; border-radius: 8px; text-decoration: none; color: #24292e; transition: all 0.2s; margin-bottom: 15px; }
		.link-item:hover { background-color: #f6f8fa; border-color: #0366d6; transform: translateY(-2px); }
		.link-title { display: block; font-size: 18px; font-weight: 600; color: #0366d6; margin-bottom: 5px; }
		.link-url { display: block; font-size: 14px; color: #586069; word-break: break-all; }
		.emoji { margin-right: 8px; }
	</style>
</head>
<body>
	<div class="container">
		<h1>Welcome to Go Clean Template API</h1>
		
		<div class="link-group">
			{{if .IsProduction}}
				<p style="color: #e67e22; font-weight: 600;">⚠️ Production Environment</p>
				<p>API documentation is not available in the production environment for security reasons.</p>
			{{else if or .SwaggerEnabled .ProtoEnabled}}
				<p>The following documentation is available for this API:</p>
				
				{{if .SwaggerEnabled}}
				<a href="{{.SwaggerURL}}" class="link-item">
					<span class="link-title"><span class="emoji">📄</span>Swagger UI</span>
					<span class="link-url">{{.SwaggerURL}}</span>
				</a>
				{{end}}
				
				{{if .ProtoEnabled}}
				<a href="{{.ProtoURL}}" class="link-item">
					<span class="link-title"><span class="emoji">🛠️</span>Protobuf Documentation</span>
					<span class="link-url">{{.ProtoURL}}</span>
				</a>
				{{end}}
			{{else}}
				<p style="color: #e74c3c; font-weight: 600;">⚠️ Documentation is currently disabled.</p>
				<p>Please check your configuration settings (SWAGGER_ENABLED/PROTO_DOCS_ENABLED) to enable API documentation.</p>
			{{end}}
		</div>
	</div>
</body>
</html>
`

func setupRoot(handler *gin.Engine, cfg *config.Config) {
	handler.GET("/", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.Request.Host

		data := struct {
			SwaggerURL     string
			ProtoURL       string
			SwaggerEnabled bool
			ProtoEnabled   bool
			IsProduction   bool
		}{
			SwaggerURL:     scheme + "://" + host + "/swagger/index.html",
			ProtoURL:       scheme + "://" + host + "/docs/proto",
			SwaggerEnabled: cfg.Swagger.Enabled,
			ProtoEnabled:   cfg.Proto.Enabled,
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
