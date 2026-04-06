package http

import (
	"net/http"

	"gct/internal/context/admin/supporting/sitesetting"
	"gct/internal/context/admin/supporting/sitesetting/application/command"
	"gct/internal/context/admin/supporting/sitesetting/application/query"
	"gct/internal/context/admin/supporting/sitesetting/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
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

// @Summary Create a site setting
// @Description Create a new site setting
// @Tags SiteSettings
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Site setting data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /site-settings [post]
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

// @Summary List site settings
// @Description Get a paginated list of site settings
// @Tags SiteSettings
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /site-settings [get]
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

// @Summary Get a site setting
// @Description Get a single site setting by ID
// @Tags SiteSettings
// @Accept json
// @Produce json
// @Param id path string true "Site Setting ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /site-settings/{id} [get]
// Get returns a single site setting by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseSiteSettingID(ctx.Param("id"))
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

// @Summary Update a site setting
// @Description Update an existing site setting by ID
// @Tags SiteSettings
// @Accept json
// @Produce json
// @Param id path string true "Site Setting ID"
// @Param request body UpdateRequest true "Site setting update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /site-settings/{id} [patch]
// Update updates a site setting.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := domain.ParseSiteSettingID(ctx.Param("id"))
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

// @Summary Delete a site setting
// @Description Delete a site setting by ID
// @Tags SiteSettings
// @Accept json
// @Produce json
// @Param id path string true "Site Setting ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /site-settings/{id} [delete]
// Delete deletes a site setting.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseSiteSettingID(ctx.Param("id"))
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
