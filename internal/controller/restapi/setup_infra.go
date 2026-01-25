package restapi

import (
	"net/http"
	"strings"

	"gct/config"
	"gct/internal/usecase"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

// setupMetrics configures the Prometheus exporter subsystem name based on app config.
func setupMetrics(handler *gin.Engine, cfg *config.Config) {
	if cfg.Metrics.Enabled {
		subsystem := strings.ReplaceAll(cfg.App.Name, "-", "_")
		subsystem = strings.ReplaceAll(subsystem, " ", "_")

		prometheus := ginprometheus.NewPrometheus(subsystem)
		prometheus.Use(handler)
	}
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
