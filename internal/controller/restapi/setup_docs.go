package restapi

import (
	"gct/config"
	docs "gct/docs/swagger"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

// setupSwagger initializes the Swagger documentation engine and dynamic host resolution.
func setupSwagger(handler *gin.Engine, cfg *config.Config) {
	docs.SwaggerInfo.Version = cfg.App.Version
	if cfg.Swagger.Enabled || cfg.App.IsDev() {
		handler.GET(swaggerRoute, func(ctx *gin.Context) {
			docs.SwaggerInfo.Host = ctx.Request.Host
			if ctx.Request.TLS != nil || httpx.GetForwardedProto(ctx) == "https" {
				docs.SwaggerInfo.Schemes = []string{"https"}
			} else {
				docs.SwaggerInfo.Schemes = []string{"http"}
			}
			ctx.Next()
		}, ginswagger.WrapHandler(swaggerfiles.Handler,
			func(c *ginswagger.Config) {
				c.Title = "Go Clean Architecture Swagger Docs"
				c.DocExpansion = "none"
				c.PersistAuthorization = true
				c.DefaultModelsExpandDepth = -1
			},
		))
	}
}

// setupProtoDocs serves generated HTML documentation for Protobuf definitions.
func setupProtoDocs(handler *gin.Engine, cfg *config.Config) {
	if cfg.Proto.Enabled || cfg.App.IsDev() {
		handler.StaticFile(protoPath, "./docs/protobuf/doc/index.html")
	}
}
