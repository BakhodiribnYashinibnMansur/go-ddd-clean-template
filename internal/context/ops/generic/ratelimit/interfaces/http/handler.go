package http

import (
	"net/http"

	"gct/internal/context/ops/generic/ratelimit"
	"gct/internal/context/ops/generic/ratelimit/application/command"
	"gct/internal/context/ops/generic/ratelimit/application/query"
	"gct/internal/context/ops/generic/ratelimit/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the RateLimit bounded context.
type Handler struct {
	bc *ratelimit.BoundedContext
	l  logger.Log
}

// NewHandler creates a new RateLimit HTTP handler.
func NewHandler(bc *ratelimit.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// @Summary Create a rate limit rule
// @Description Create a new rate limit rule
// @Tags RateLimits
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Rate limit data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /rate-limits [post]
// Create creates a new rate limit rule.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateRateLimitCommand{
		Name:              req.Name,
		Rule:              req.Rule,
		RequestsPerWindow: req.RequestsPerWindow,
		WindowDuration:    req.WindowDuration,
		Enabled:           req.Enabled,
	}
	if err := h.bc.CreateRateLimit.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary List rate limit rules
// @Description Return a paginated list of rate limit rules
// @Tags RateLimits
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /rate-limits [get]
// List returns a paginated list of rate limit rules.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListRateLimitsQuery{
		Filter: domain.RateLimitFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListRateLimits.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.RateLimits, "total": result.Total})
}

// @Summary Get a rate limit rule
// @Description Return a single rate limit rule by ID
// @Tags RateLimits
// @Accept json
// @Produce json
// @Param id path string true "Rate limit ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /rate-limits/{id} [get]
// Get returns a single rate limit rule by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseRateLimitID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetRateLimit.Handle(ctx.Request.Context(), query.GetRateLimitQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Update a rate limit rule
// @Description Update an existing rate limit rule
// @Tags RateLimits
// @Accept json
// @Produce json
// @Param id path string true "Rate limit ID"
// @Param request body UpdateRequest true "Rate limit update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /rate-limits/{id} [patch]
// Update updates a rate limit rule.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := domain.ParseRateLimitID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.UpdateRateLimitCommand{
		ID:                domain.RateLimitID(id),
		Name:              req.Name,
		Rule:              req.Rule,
		RequestsPerWindow: req.RequestsPerWindow,
		WindowDuration:    req.WindowDuration,
		Enabled:           req.Enabled,
	}
	if err := h.bc.UpdateRateLimit.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Delete a rate limit rule
// @Description Delete a rate limit rule by ID
// @Tags RateLimits
// @Accept json
// @Produce json
// @Param id path string true "Rate limit ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /rate-limits/{id} [delete]
// Delete deletes a rate limit rule.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseRateLimitID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteRateLimit.Handle(ctx.Request.Context(), command.DeleteRateLimitCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
