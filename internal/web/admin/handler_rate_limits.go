package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RateLimits (GET) - List rate limit rules
func (h *Handler) RateLimits(ctx *gin.Context) {
	h.servePage(ctx, "rate_limits/list.html", "Rate Limits", "rate_limits", map[string]any{
		"Rules":          []any{},
		"TotalCount":     0,
		"ActiveCount":    0,
		"BlockedCount":   0,
		"HitsToday":      0,
		"Scopes":         []string{"GLOBAL", "ENDPOINT", "USER", "IP", "API_KEY"},
		"Actions":        []string{"REJECT", "THROTTLE", "BLOCK", "LOG_ONLY"},
	})
}

// CreateRateLimitPost (POST)
func (h *Handler) CreateRateLimitPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — rate limits backend not connected"})
}

// UpdateRateLimitPost (PUT)
func (h *Handler) UpdateRateLimitPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — rate limits backend not connected"})
}

// DeleteRateLimitPost (DELETE)
func (h *Handler) DeleteRateLimitPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — rate limits backend not connected"})
}

// ToggleRateLimitPost (POST)
func (h *Handler) ToggleRateLimitPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — rate limits backend not connected"})
}
