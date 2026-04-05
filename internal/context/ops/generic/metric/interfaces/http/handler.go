package http

import (
	"net/http"

	"gct/internal/context/ops/generic/metric"
	"gct/internal/context/ops/generic/metric/application/command"
	"gct/internal/context/ops/generic/metric/application/query"
	"gct/internal/context/ops/generic/metric/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for the Metric bounded context.
type Handler struct {
	bc *metric.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Metric HTTP handler.
func NewHandler(bc *metric.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create records a new metric.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.RecordMetricCommand{
		Name:       req.Name,
		LatencyMs:  req.LatencyMs,
		IsPanic:    req.IsPanic,
		PanicError: req.PanicError,
	}
	if err := h.bc.RecordMetric.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of metrics.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListMetricsQuery{
		Filter: domain.MetricFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListMetrics.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Metrics, "total": result.Total})
}
