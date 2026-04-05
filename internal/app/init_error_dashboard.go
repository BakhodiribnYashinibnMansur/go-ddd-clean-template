package app

import (
	"net/http"

	apperrors "gct/internal/platform/infrastructure/errors"

	"github.com/gin-gonic/gin"
)

func registerErrorDashboardRoutes(r *gin.RouterGroup) {
	r.GET("/errors/stats", handleErrorStats)
}

func handleErrorStats(c *gin.Context) {
	metrics := apperrors.GetGlobalMetrics()
	stats := metrics.GetStats()
	c.JSON(http.StatusOK, stats)
}
