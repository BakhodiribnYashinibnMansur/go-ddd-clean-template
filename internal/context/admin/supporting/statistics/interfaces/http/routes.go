package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all Statistics HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	s := rg.Group("/statistics")
	{
		s.GET("/overview", h.GetOverview)
		s.GET("/users", h.GetUserStats)
		s.GET("/sessions", h.GetSessionStats)
		s.GET("/errors", h.GetErrorStats)
		s.GET("/audit", h.GetAuditStats)
		s.GET("/security", h.GetSecurityStats)
		s.GET("/feature-flags", h.GetFeatureFlagStats)
		s.GET("/content", h.GetContentStats)
		s.GET("/integrations", h.GetIntegrationStats)
	}
}
