package admin

import (
	"gct/pkg/httpx"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *Handler) FunctionMetrics(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.FunctionMetricsFilter{
		Pagination: pagination,
	}

	if name := httpx.GetNullStringQuery(ctx, "name"); name != "" {
		filter.Name = &name
	}
	if isPanicStr := httpx.GetNullStringQuery(ctx, "is_panic"); isPanicStr != "" {
		p := isPanicStr == "true"
		filter.IsPanic = &p
	}

	metrics, count, err := h.uc.Audit.Metric.Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch function metrics", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "functions.html", "Function Metrics", "functions", map[string]any{
		"Metrics":     metrics,
		"Pagination":  pagination,
		"Filter":      filter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}
