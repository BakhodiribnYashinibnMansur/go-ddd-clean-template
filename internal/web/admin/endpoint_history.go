package admin

import (
	"strconv"

	"gct/pkg/httpx"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) EndpointHistory(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.EndpointHistoriesFilter{
		Pagination: pagination,
	}

	if userID := httpx.GetNullStringQuery(ctx, "user_id"); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			filter.UserID = &uid
		}
	}
	if method := httpx.GetNullStringQuery(ctx, "method"); method != "" {
		filter.Method = &method
	}
	if path := httpx.GetNullStringQuery(ctx, "path"); path != "" {
		filter.Path = &path
	}
	if statusStr := httpx.GetNullStringQuery(ctx, "status_code"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			filter.StatusCode = &status
		}
	}

	history, count, err := h.uc.Audit.History.Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch endpoint history", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "audit/history.html", "Endpoint History", "endpoint_history", map[string]any{
		"History":     history,
		"Pagination":  pagination,
		"Filter":      filter.EndpointHistoryFilter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}
