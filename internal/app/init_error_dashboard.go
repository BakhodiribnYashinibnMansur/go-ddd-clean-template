package app

import (
	"net/http"

	apperrors "gct/internal/kernel/infrastructure/errorx"

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
