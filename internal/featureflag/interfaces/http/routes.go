package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all FeatureFlag HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/feature-flags")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", h.Update)
	g.DELETE("/:id", h.Delete)

	// Rule group sub-routes
	g.POST("/:id/rule-groups", h.CreateRuleGroup)
	g.PATCH("/:id/rule-groups/:groupId", h.UpdateRuleGroup)
	g.DELETE("/:id/rule-groups/:groupId", h.DeleteRuleGroup)
}
