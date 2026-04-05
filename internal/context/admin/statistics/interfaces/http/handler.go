package http

import (
	"net/http"

	"gct/internal/context/admin/statistics"
	"gct/internal/context/admin/statistics/application/query"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the Statistics bounded context.
type Handler struct {
	bc *statistics.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Statistics HTTP handler.
func NewHandler(bc *statistics.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// GetOverview returns the top-level statistics aggregate.
func (h *Handler) GetOverview(ctx *gin.Context) {
	result, err := h.bc.GetOverview.Handle(ctx.Request.Context(), query.GetOverviewQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetUserStats returns the user lifecycle and role breakdown.
func (h *Handler) GetUserStats(ctx *gin.Context) {
	result, err := h.bc.GetUserStats.Handle(ctx.Request.Context(), query.GetUserStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetSessionStats returns the session state breakdown.
func (h *Handler) GetSessionStats(ctx *gin.Context) {
	result, err := h.bc.GetSessionStats.Handle(ctx.Request.Context(), query.GetSessionStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetErrorStats returns the system error breakdown.
func (h *Handler) GetErrorStats(ctx *gin.Context) {
	result, err := h.bc.GetErrorStats.Handle(ctx.Request.Context(), query.GetErrorStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetAuditStats returns the audit log recency breakdown.
func (h *Handler) GetAuditStats(ctx *gin.Context) {
	result, err := h.bc.GetAuditStats.Handle(ctx.Request.Context(), query.GetAuditStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetSecurityStats returns IP rule and rate limit counts.
func (h *Handler) GetSecurityStats(ctx *gin.Context) {
	result, err := h.bc.GetSecurityStats.Handle(ctx.Request.Context(), query.GetSecurityStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetFeatureFlagStats returns the feature flag active-state breakdown.
func (h *Handler) GetFeatureFlagStats(ctx *gin.Context) {
	result, err := h.bc.GetFeatureFlagStats.Handle(ctx.Request.Context(), query.GetFeatureFlagStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetContentStats returns content table counts.
func (h *Handler) GetContentStats(ctx *gin.Context) {
	result, err := h.bc.GetContentStats.Handle(ctx.Request.Context(), query.GetContentStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// GetIntegrationStats returns integration and API key counts.
func (h *Handler) GetIntegrationStats(ctx *gin.Context) {
	result, err := h.bc.GetIntegrationStats.Handle(ctx.Request.Context(), query.GetIntegrationStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}
