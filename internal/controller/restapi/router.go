// Package restapi implements routing paths. Each services in own file.
package restapi

import (
	"net/http"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/docs" // Swagger docs.
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/user"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
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
// @BasePath    /v1
func NewRouter(handler *gin.Engine, cfg *config.Config, uc *usecase.UseCase, l logger.Log) {
	// Options
	handler.HandleMethodNotAllowed = true

	middleware.GinMiddleware(handler)

	// Prometheus metrics
	setupMetrics(handler, cfg)

	// Swagger settings
	setupSwagger(handler, cfg)

	// K8s probe
	setupHealthCheck(handler)

	// Controller
	c := NewController(uc, cfg, l)

	// Middleware
	am := middleware.NewAuthMiddleware(uc, cfg, l)

	// Routers
	h := handler.Group("/v1")
	{
		user.UserRoute(h, c.User, am.AuthClientAccess)
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

func setupHealthCheck(handler *gin.Engine) {
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	handler.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })

}
