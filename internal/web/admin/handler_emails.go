package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// EmailTemplates (GET) - List email templates
func (h *Handler) EmailTemplates(ctx *gin.Context) {
	h.servePage(ctx, "emails/templates.html", "Email Templates", "emails", map[string]any{
		"Templates":    []any{},
		"TotalCount":   0,
		"ActiveCount":  0,
		"Categories":   []string{"general", "auth", "notification", "marketing", "transactional"},
	})
}

// EmailLogs (GET) - List email logs
func (h *Handler) EmailLogs(ctx *gin.Context) {
	h.servePage(ctx, "emails/logs.html", "Email Logs", "emails", map[string]any{
		"Logs":         []any{},
		"TotalCount":   0,
		"SentToday":    0,
		"FailedCount":  0,
		"BounceRate":   0.0,
	})
}

// CreateEmailTemplatePost (POST)
func (h *Handler) CreateEmailTemplatePost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — email backend not connected"})
}

// UpdateEmailTemplatePost (PUT)
func (h *Handler) UpdateEmailTemplatePost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — email backend not connected"})
}

// DeleteEmailTemplatePost (DELETE)
func (h *Handler) DeleteEmailTemplatePost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — email backend not connected"})
}

// SendTestEmailPost (POST)
func (h *Handler) SendTestEmailPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — email backend not connected"})
}
