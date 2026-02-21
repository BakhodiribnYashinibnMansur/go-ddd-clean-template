package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Media (GET) - Media library
func (h *Handler) Media(ctx *gin.Context) {
	h.servePage(ctx, "media/list.html", "Media Library", "media", map[string]any{
		"Files":        []any{},
		"TotalCount":   0,
		"TotalSize":    "0 B",
		"ImageCount":   0,
		"DocCount":     0,
	})
}

// UploadMediaPost (POST)
func (h *Handler) UploadMediaPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — media backend not connected"})
}

// UpdateMediaPost (PUT)
func (h *Handler) UpdateMediaPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — media backend not connected"})
}

// DeleteMediaPost (DELETE)
func (h *Handler) DeleteMediaPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — media backend not connected"})
}
