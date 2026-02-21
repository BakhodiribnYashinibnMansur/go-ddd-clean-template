package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Webhooks (GET) - List webhooks
func (h *Handler) Webhooks(ctx *gin.Context) {
	h.servePage(ctx, "webhooks/list.html", "Webhooks", "webhooks", map[string]any{
		"Webhooks":    []any{},
		"TotalCount":  0,
		"ActiveCount": 0,
		"FailedCount": 0,
		"Events": []string{
			"user.created", "user.updated", "user.deleted", "user.blocked",
			"session.created", "session.revoked",
			"role.created", "role.updated", "role.deleted",
			"permission.created", "permission.updated", "permission.deleted",
			"policy.created", "policy.updated", "policy.deleted",
			"integration.created", "integration.updated", "integration.deleted",
			"audit.access_denied", "audit.login", "audit.logout",
			"system.error", "system.health_check_failed",
		},
	})
}

// WebhookDetail (GET)
func (h *Handler) WebhookDetail(ctx *gin.Context) {
	h.servePage(ctx, "webhooks/detail.html", "Webhook Detail", "webhooks", map[string]any{
		"Webhook":    nil,
		"Deliveries": []any{},
	})
}

// CreateWebhookPost (POST)
func (h *Handler) CreateWebhookPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — webhooks backend not connected"})
}

// UpdateWebhookPost (PUT)
func (h *Handler) UpdateWebhookPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — webhooks backend not connected"})
}

// DeleteWebhookPost (DELETE)
func (h *Handler) DeleteWebhookPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — webhooks backend not connected"})
}

// TestWebhookPost (POST) - Send test payload
func (h *Handler) TestWebhookPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — webhooks backend not connected"})
}
