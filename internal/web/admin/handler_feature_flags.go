package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// FeatureFlags (GET) - List feature flags
func (h *Handler) FeatureFlags(ctx *gin.Context) {
	h.servePage(ctx, "feature_flags/list.html", "Feature Flags", "feature_flags", map[string]any{
		"Flags":        []any{},
		"TotalCount":   0,
		"EnabledCount": 0,
		"Categories":   []string{"general", "ui", "backend", "experimental"},
	})
}

// CreateFeatureFlagPost (POST)
func (h *Handler) CreateFeatureFlagPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — feature flags backend not connected"})
}

// UpdateFeatureFlagPost (PUT)
func (h *Handler) UpdateFeatureFlagPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — feature flags backend not connected"})
}

// DeleteFeatureFlagPost (DELETE)
func (h *Handler) DeleteFeatureFlagPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — feature flags backend not connected"})
}

// ToggleFeatureFlagPost (POST)
func (h *Handler) ToggleFeatureFlagPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — feature flags backend not connected"})
}
