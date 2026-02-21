package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// IPRules (GET) - List IP rules
func (h *Handler) IPRules(ctx *gin.Context) {
	h.servePage(ctx, "ip_rules/list.html", "IP Rules", "ip_rules", map[string]any{
		"Rules":        []any{},
		"TotalCount":   0,
		"AllowCount":   0,
		"DenyCount":    0,
		"AutoBlocked":  0,
		"Scopes":       []string{"ALL", "ADMIN_PANEL", "API", "PUBLIC"},
	})
}

// CreateIPRulePost (POST)
func (h *Handler) CreateIPRulePost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — IP rules backend not connected"})
}

// UpdateIPRulePost (PUT)
func (h *Handler) UpdateIPRulePost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — IP rules backend not connected"})
}

// DeleteIPRulePost (DELETE)
func (h *Handler) DeleteIPRulePost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — IP rules backend not connected"})
}
