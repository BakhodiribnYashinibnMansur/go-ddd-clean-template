package http

import (
	"net/http"

	"gct/internal/context/admin/supporting/integration"
	"gct/internal/context/admin/supporting/integration/application/command"
	"gct/internal/context/admin/supporting/integration/application/query"
	"gct/internal/context/admin/supporting/integration/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the Integration bounded context.
type Handler struct {
	bc *integration.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Integration HTTP handler.
func NewHandler(bc *integration.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// @Summary Create an integration
// @Description Create a new external integration
// @Tags Integrations
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Integration data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /integrations [post]
// Create creates a new integration.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateCommand{
		Name:       req.Name,
		Type:       req.Type,
		APIKey:     req.APIKey,
		WebhookURL: req.WebhookURL,
		Enabled:    req.Enabled,
		Config:     req.Config,
	}
	if err := h.bc.CreateIntegration.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary List integrations
// @Description Get a paginated list of integrations
// @Tags Integrations
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /integrations [get]
// List returns a paginated list of integrations.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListQuery{
		Filter: domain.IntegrationFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListIntegrations.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Integrations, "total": result.Total})
}

// @Summary Get an integration
// @Description Get a single integration by ID
// @Tags Integrations
// @Accept json
// @Produce json
// @Param id path string true "Integration ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /integrations/{id} [get]
// Get returns a single integration by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseIntegrationID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetIntegration.Handle(ctx.Request.Context(), query.GetQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Update an integration
// @Description Update an existing integration by ID
// @Tags Integrations
// @Accept json
// @Produce json
// @Param id path string true "Integration ID"
// @Param request body UpdateRequest true "Integration update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /integrations/{id} [patch]
// Update updates an integration.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := domain.ParseIntegrationID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.UpdateCommand{
		ID:         id,
		Name:       req.Name,
		Type:       req.Type,
		APIKey:     req.APIKey,
		WebhookURL: req.WebhookURL,
		Enabled:    req.Enabled,
		Config:     req.Config,
	}
	if err := h.bc.UpdateIntegration.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Delete an integration
// @Description Delete an integration by ID
// @Tags Integrations
// @Accept json
// @Produce json
// @Param id path string true "Integration ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /integrations/{id} [delete]
// Delete deletes an integration.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseIntegrationID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteIntegration.Handle(ctx.Request.Context(), command.DeleteCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
