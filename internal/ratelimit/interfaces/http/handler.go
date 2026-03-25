package http

import (
	"net/http"
	"strconv"

	"gct/internal/ratelimit"
	"gct/internal/ratelimit/application/command"
	"gct/internal/ratelimit/application/query"
	"gct/internal/ratelimit/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// Create creates a new rate limit rule.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of rate limit rules.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListRateLimitsQuery{
		Filter: domain.RateLimitFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListRateLimits.Handle(ctx.Request.Context(), q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.RateLimits, "total": result.Total})
}

// Get returns a single rate limit rule by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.bc.GetRateLimit.Handle(ctx.Request.Context(), query.GetRateLimitQuery{ID: id})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Update updates a rate limit rule.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cmd := command.UpdateRateLimitCommand{
		ID:                id,
		Name:              req.Name,
		Rule:              req.Rule,
		RequestsPerWindow: req.RequestsPerWindow,
		WindowDuration:    req.WindowDuration,
		Enabled:           req.Enabled,
	}
	if err := h.bc.UpdateRateLimit.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete deletes a rate limit rule.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.bc.DeleteRateLimit.Handle(ctx.Request.Context(), command.DeleteRateLimitCommand{ID: id}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
