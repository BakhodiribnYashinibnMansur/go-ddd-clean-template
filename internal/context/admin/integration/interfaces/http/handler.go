package http

import (
	"net/http"
	"strconv"

	"gct/internal/context/admin/integration"
	"gct/internal/context/admin/integration/application/command"
	"gct/internal/context/admin/integration/application/query"
	"gct/internal/context/admin/integration/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// List returns a paginated list of integrations.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListQuery{
		Filter: domain.IntegrationFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListIntegrations.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Integrations, "total": result.Total})
}

// Get returns a single integration by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetIntegration.Handle(ctx.Request.Context(), query.GetQuery{ID: domain.IntegrationID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Update updates an integration.
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
	cmd := command.UpdateCommand{
		ID:         domain.IntegrationID(id),
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

// Delete deletes an integration.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteIntegration.Handle(ctx.Request.Context(), command.DeleteCommand{ID: domain.IntegrationID(id)}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
