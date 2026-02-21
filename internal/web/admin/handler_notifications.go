package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Notifications (GET) - List notifications
func (h *Handler) Notifications(ctx *gin.Context) {
	h.servePage(ctx, "notifications/list.html", "Notifications", "notifications", map[string]any{
		"Notifications": []any{},
		"TotalCount":    0,
		"UnreadCount":   0,
		"SentToday":     0,
		"FailedCount":   0,
	})
}

// CreateNotificationPost (POST)
func (h *Handler) CreateNotificationPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — notifications backend not connected"})
}

// UpdateNotificationPost (PUT)
func (h *Handler) UpdateNotificationPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — notifications backend not connected"})
}

// DeleteNotificationPost (DELETE)
func (h *Handler) DeleteNotificationPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — notifications backend not connected"})
}
