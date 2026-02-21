package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DataExport (GET) - Data export/import center
func (h *Handler) DataExport(ctx *gin.Context) {
	h.servePage(ctx, "data_export/list.html", "Data Export", "data_export", map[string]any{
		"Exports":      []any{},
		"Imports":      []any{},
		"ExportCount":  0,
		"ImportCount":  0,
		"Resources": []string{
			"users", "sessions", "roles", "permissions", "policies",
			"audit_log", "endpoint_history", "system_errors", "error_codes",
			"integrations", "relations",
		},
		"Formats": []string{"CSV", "XLSX", "JSON", "PDF"},
	})
}

// CreateExportPost (POST)
func (h *Handler) CreateExportPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — data export backend not connected"})
}

// DeleteExportPost (DELETE)
func (h *Handler) DeleteExportPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — data export backend not connected"})
}
