package http

import (
	"net/http"
	"strconv"

	"gct/internal/context/ops/metric"
	"gct/internal/context/ops/metric/application/command"
	"gct/internal/context/ops/metric/application/query"
	"gct/internal/context/ops/metric/domain"
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
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListMetricsQuery{
		Filter: domain.MetricFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListMetrics.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Metrics, "total": result.Total})
}
