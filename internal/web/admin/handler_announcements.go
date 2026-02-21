package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Announcements (GET) - List announcements
func (h *Handler) Announcements(ctx *gin.Context) {
	h.servePage(ctx, "announcements/list.html", "Announcements", "announcements", map[string]any{
		"Announcements": []any{},
		"TotalCount":    0,
		"ActiveCount":   0,
		"ScheduledCount": 0,
		"Types":         []string{"INFO", "WARNING", "CRITICAL", "SUCCESS", "PROMOTION"},
		"Locations":     []string{"TOP_BAR", "DASHBOARD", "LOGIN_PAGE", "SIDEBAR", "MODAL"},
	})
}

// CreateAnnouncementPost (POST)
func (h *Handler) CreateAnnouncementPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — announcements backend not connected"})
}

// UpdateAnnouncementPost (PUT)
func (h *Handler) UpdateAnnouncementPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — announcements backend not connected"})
}

// DeleteAnnouncementPost (DELETE)
func (h *Handler) DeleteAnnouncementPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — announcements backend not connected"})
}

// ToggleAnnouncementPost (POST)
func (h *Handler) ToggleAnnouncementPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — announcements backend not connected"})
}
