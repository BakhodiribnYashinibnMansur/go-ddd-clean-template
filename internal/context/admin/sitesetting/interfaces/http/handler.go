package http

import (
	"net/http"
	"strconv"

	"gct/internal/platform/infrastructure/httpx"
	"gct/internal/platform/infrastructure/httpx/response"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/context/admin/sitesetting"
	"gct/internal/context/admin/sitesetting/application/command"
	"gct/internal/context/admin/sitesetting/application/query"
	"gct/internal/context/admin/sitesetting/domain"

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
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListSiteSettingsQuery{
		Filter: domain.SiteSettingFilter{Limit: limit, Offset: offset},
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
	result, err := h.bc.GetSiteSetting.Handle(ctx.Request.Context(), query.GetSiteSettingQuery{ID: id})
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
		ID:          id,
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
	if err := h.bc.DeleteSiteSetting.Handle(ctx.Request.Context(), command.DeleteSiteSettingCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
