package admin

import (
	"github.com/gin-gonic/gin"
)

// APIAnalytics (GET) - API analytics dashboard
func (h *Handler) APIAnalytics(ctx *gin.Context) {
	h.servePage(ctx, "api_analytics/dashboard.html", "API Analytics", "api_analytics", map[string]any{
		"TotalCallsToday":   0,
		"ActiveIntegrations": 0,
		"AvgResponseTime":   "0ms",
		"ErrorRate":         "0%",
		"TopEndpoints":      []any{},
		"RecentErrors":      []any{},
	})
}
