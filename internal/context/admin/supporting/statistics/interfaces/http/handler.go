package http

import (
	"net/http"

	"gct/internal/context/admin/supporting/statistics"
	"gct/internal/context/admin/supporting/statistics/application/query"
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

// @Summary Get overview statistics
// @Description Returns the top-level statistics aggregate
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /statistics/overview [get]
// GetOverview returns the top-level statistics aggregate.
func (h *Handler) GetOverview(ctx *gin.Context) {
	result, err := h.bc.GetOverview.Handle(ctx.Request.Context(), query.GetOverviewQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Get user statistics
// @Description Returns the user lifecycle and role breakdown
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /statistics/users [get]
// GetUserStats returns the user lifecycle and role breakdown.
func (h *Handler) GetUserStats(ctx *gin.Context) {
	result, err := h.bc.GetUserStats.Handle(ctx.Request.Context(), query.GetUserStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Get session statistics
// @Description Returns the session state breakdown
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /statistics/sessions [get]
// GetSessionStats returns the session state breakdown.
func (h *Handler) GetSessionStats(ctx *gin.Context) {
	result, err := h.bc.GetSessionStats.Handle(ctx.Request.Context(), query.GetSessionStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Get error statistics
// @Description Returns the system error breakdown
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /statistics/errors [get]
// GetErrorStats returns the system error breakdown.
func (h *Handler) GetErrorStats(ctx *gin.Context) {
	result, err := h.bc.GetErrorStats.Handle(ctx.Request.Context(), query.GetErrorStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Get audit statistics
// @Description Returns the audit log recency breakdown
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /statistics/audit [get]
// GetAuditStats returns the audit log recency breakdown.
func (h *Handler) GetAuditStats(ctx *gin.Context) {
	result, err := h.bc.GetAuditStats.Handle(ctx.Request.Context(), query.GetAuditStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Get security statistics
// @Description Returns IP rule and rate limit counts
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /statistics/security [get]
// GetSecurityStats returns IP rule and rate limit counts.
func (h *Handler) GetSecurityStats(ctx *gin.Context) {
	result, err := h.bc.GetSecurityStats.Handle(ctx.Request.Context(), query.GetSecurityStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Get feature flag statistics
// @Description Returns the feature flag active-state breakdown
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /statistics/feature-flags [get]
// GetFeatureFlagStats returns the feature flag active-state breakdown.
func (h *Handler) GetFeatureFlagStats(ctx *gin.Context) {
	result, err := h.bc.GetFeatureFlagStats.Handle(ctx.Request.Context(), query.GetFeatureFlagStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Get content statistics
// @Description Returns content table counts
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /statistics/content [get]
// GetContentStats returns content table counts.
func (h *Handler) GetContentStats(ctx *gin.Context) {
	result, err := h.bc.GetContentStats.Handle(ctx.Request.Context(), query.GetContentStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Get integration statistics
// @Description Returns integration and API key counts
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /statistics/integrations [get]
// GetIntegrationStats returns integration and API key counts.
func (h *Handler) GetIntegrationStats(ctx *gin.Context) {
	result, err := h.bc.GetIntegrationStats.Handle(ctx.Request.Context(), query.GetIntegrationStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}
