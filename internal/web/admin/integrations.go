package admin

import (
	"net/http"

	"gct/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) Integrations(ctx *gin.Context) {
	filter := domain.IntegrationFilter{
		Limit:  100,
		Offset: 0,
	}

	if search := ctx.Query("search"); search != "" {
		filter.Search = search
	}
	if status := ctx.Query("status"); status != "" {
		active := status == "active"
		filter.IsActive = &active
	}

	integrations, count, err := h.uc.Integration.ListIntegrations(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch integrations", "error", err)
	}

	h.servePage(ctx, "integrations/list.html", "Integrations", "integrations", map[string]any{
		"Integrations":  integrations,
		"TotalCount":    count,
		"FilterSearch":  ctx.Query("search"),
		"FilterStatus":  ctx.Query("status"),
		"QueryParams":   ctx.Request.URL.Query(),
	})
}

func (h *Handler) IntegrationDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/admin/integrations")
		return
	}

	integration, err := h.uc.Integration.GetIntegration(ctx.Request.Context(), id)
	if err != nil {
		h.l.Errorw("failed to fetch integration", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/integrations")
		return
	}

	h.servePage(ctx, "integrations/detail.html", "Integration Details", "integrations", map[string]any{
		"Integration": integration,
	})
}

func (h *Handler) CreateIntegrationPost(ctx *gin.Context) {
	var req domain.CreateIntegrationRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	integration, err := h.uc.Integration.CreateIntegration(ctx.Request.Context(), req)
	if err != nil {
		h.l.Errorw("failed to create integration", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Integration created", "id": integration.ID})
}

func (h *Handler) UpdateIntegrationPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	var req domain.UpdateIntegrationRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	_, err = h.uc.Integration.UpdateIntegration(ctx.Request.Context(), id, req)
	if err != nil {
		h.l.Errorw("failed to update integration", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Integration updated"})
}

func (h *Handler) DeleteIntegrationPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	err = h.uc.Integration.DeleteIntegration(ctx.Request.Context(), id)
	if err != nil {
		h.l.Errorw("failed to delete integration", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Integration deleted"})
}

func (h *Handler) ToggleIntegrationPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	integration, err := h.uc.Integration.GetIntegration(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Integration not found"})
		return
	}

	newActive := !integration.IsActive
	_, err = h.uc.Integration.UpdateIntegration(ctx.Request.Context(), id, domain.UpdateIntegrationRequest{
		IsActive: &newActive,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Integration toggled", "is_active": newActive})
}

func (h *Handler) CreateAPIKeyPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	integrationID, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid integration ID"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	apiKey, rawKey, err := h.uc.Integration.CreateAPIKey(ctx.Request.Context(), domain.CreateAPIKeyRequest{
		IntegrationID: integrationID,
		Name:          req.Name,
	})
	if err != nil {
		h.l.Errorw("failed to create API key", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API key created",
		"key":     rawKey,
		"id":      apiKey.ID,
		"prefix":  apiKey.KeyPrefix,
	})
}

func (h *Handler) RevokeAPIKeyPost(ctx *gin.Context) {
	kidStr := ctx.Param("kid")
	kid, err := uuid.Parse(kidStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid key ID"})
		return
	}

	err = h.uc.Integration.RevokeAPIKey(ctx.Request.Context(), kid)
	if err != nil {
		h.l.Errorw("failed to revoke API key", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "API key revoked"})
}

func (h *Handler) DeleteAPIKeyPost(ctx *gin.Context) {
	kidStr := ctx.Param("kid")
	kid, err := uuid.Parse(kidStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid key ID"})
		return
	}

	err = h.uc.Integration.DeleteAPIKey(ctx.Request.Context(), kid)
	if err != nil {
		h.l.Errorw("failed to delete API key", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "API key deleted"})
}
