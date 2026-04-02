package http

import (
	"net/http"

	"gct/internal/dashboard"
	"gct/internal/dashboard/application/query"
	"gct/internal/shared/infrastructure/httpx/response"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the Dashboard bounded context.
type Handler struct {
	bc *dashboard.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Dashboard HTTP handler.
func NewHandler(bc *dashboard.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// GetStats returns dashboard statistics.
func (h *Handler) GetStats(ctx *gin.Context) {
	result, err := h.bc.GetStats.Handle(ctx.Request.Context(), query.GetStatsQuery{})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}
