package http

import (
	"net/http"

	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/context/admin/supporting/sitesetting"
	"gct/internal/context/admin/supporting/sitesetting/application/command"
	"gct/internal/context/admin/supporting/sitesetting/application/query"
	"gct/internal/context/admin/supporting/sitesetting/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP endpoints for the SiteSetting bounded context.
type Handler struct {
	bc *sitesetting.BoundedContext
	l  logger.Log
}

// NewHandler creates a new SiteSetting HTTP handler.
func NewHandler(bc *sitesetting.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create creates a new site setting.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateSiteSettingCommand{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Description: req.Description,
	}
	if err := h.bc.CreateSiteSetting.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of site settings.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListSiteSettingsQuery{
		Filter: domain.SiteSettingFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListSiteSettings.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Settings, "total": result.Total})
}

// Get returns a single site setting by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetSiteSetting.Handle(ctx.Request.Context(), query.GetSiteSettingQuery{ID: domain.SiteSettingID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Update updates a site setting.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.UpdateSiteSettingCommand{
		ID:          domain.SiteSettingID(id),
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Description: req.Description,
	}
	if err := h.bc.UpdateSiteSetting.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete deletes a site setting.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteSiteSetting.Handle(ctx.Request.Context(), command.DeleteSiteSettingCommand{ID: domain.SiteSettingID(id)}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
