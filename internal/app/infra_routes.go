package app

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"time"

	"gct/config"
	docs "gct/docs/swagger"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	miniogo "github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

// setupInfraRoutes registers infrastructure endpoints (metrics, docs, health, static).
func setupInfraRoutes(handler *gin.Engine, cfg *config.Config, pool *pgxpool.Pool, redisClient *redis.Client, metricsHandler http.Handler, minioClient *miniogo.Client) {
	// Prometheus metrics via OTel exporter
	if cfg.Middleware.Metrics && cfg.Metrics.Enabled && metricsHandler != nil {
		handler.GET("/metrics", gin.WrapH(metricsHandler))
	}

	// Swagger
	setupSwaggerRoutes(handler, cfg)

	// Proto docs
	if cfg.Proto.Enabled || cfg.App.IsDev() {
		handler.StaticFile("/docs/proto", "./docs/protobuf/doc/index.html")
	}

	// Root landing page
	setupRootPage(handler, cfg)

	// Health checks (DDD: direct pool/redis ping, no usecase)
	if cfg.Middleware.HealthCheck {
		handler.GET("/health/live", func(c *gin.Context) { c.Status(http.StatusOK) })
		handler.GET("/health/ready", func(c *gin.Context) {
			checks := make(map[string]string)
			healthy := true

			// PostgreSQL
			if err := pool.Ping(c.Request.Context()); err != nil {
				checks["postgres"] = err.Error()
				healthy = false
			} else {
				checks["postgres"] = "ok"
			}

			// Redis
			if redisClient != nil {
				if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
					checks["redis"] = err.Error()
					healthy = false
				} else {
					checks["redis"] = "ok"
				}
			}

			// MinIO
			if minioClient != nil && cfg.Minio.Enabled {
				checkCtx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
				defer cancel()
				if _, err := minioClient.BucketExists(checkCtx, cfg.Minio.Bucket); err != nil {
					checks["minio"] = err.Error()
					healthy = false
				} else {
					checks["minio"] = "ok"
				}
			}

			if !healthy {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "checks": checks})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ready", "checks": checks})
		})
		handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
		handler.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	}

	// Browser/static routes
	handler.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})
	handler.GET("/favicon.ico", func(c *gin.Context) { c.Status(204) })
	handler.Static("/docs/linter", "./docs/report/linter")

	// Admin redirect
	setupAdminRedirectPage(handler)
}

const swaggerRoute = "/docs/swagger/*any"
const swaggerPath = "/docs/swagger/index.html"
const protoPath = "/docs/proto"
const adminPath = "/admin/dashboard"

func setupSwaggerRoutes(handler *gin.Engine, cfg *config.Config) {
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

const rootHTML = `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><title>Go Clean Template API</title>
<style>body{font-family:sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;background:#030712;color:#f8fafc}
.c{text-align:center;max-width:600px;padding:2rem}.h{font-size:3rem;margin-bottom:1rem}
a{color:#38bdf8;text-decoration:none;margin:0 1rem}</style></head>
<body><div class="c"><div class="h">Go Clean Template</div>
<p>API is running</p>
<p><a href="{{.SwaggerURL}}">Swagger</a> | <a href="{{.ProtoURL}}">Proto</a> | <a href="{{.AdminURL}}">Admin</a></p>
</div></body></html>`

func setupRootPage(handler *gin.Engine, cfg *config.Config) {
	handler.GET("/", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || httpx.GetForwardedProto(c) == "https" {
			scheme = "https"
		}
		data := struct {
			SwaggerURL, ProtoURL, AdminURL string
		}{
			SwaggerURL: scheme + "://" + c.Request.Host + swaggerPath,
			ProtoURL:   scheme + "://" + c.Request.Host + protoPath,
			AdminURL:   scheme + "://" + c.Request.Host + adminPath,
		}
		tmpl, _ := template.New("root").Parse(rootHTML)
		var buf bytes.Buffer
		_ = tmpl.Execute(&buf, data)
		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	})
}

const reactAdminURL = "http://localhost:3000"

func setupAdminRedirectPage(handler *gin.Engine) {
	handler.GET("/admin", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, reactAdminURL)
	})
	handler.GET("/admin/*path", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, reactAdminURL)
	})
}
