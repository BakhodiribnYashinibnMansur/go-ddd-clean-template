package admin

import (
	"gct/internal/controller/restapi/util"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SystemErrors(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.SystemErrorsFilter{
		Pagination: pagination,
	}

	// Filters
	if code := util.GetNullStringQuery(ctx, "code"); code != "" {
		filter.Code = &code
	}
	if severity := util.GetNullStringQuery(ctx, "severity"); severity != "" {
		filter.Severity = &severity
	}
	if isResolvedStr := util.GetNullStringQuery(ctx, "is_resolved"); isResolvedStr != "" {
		resolved := isResolvedStr == "true"
		filter.IsResolved = &resolved
	}

	errors, count, err := h.uc.Audit.SystemError.Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch system errors", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "system_errors.html", "System Errors", "system_errors", map[string]any{
		"Errors":      errors,
		"Pagination":  pagination,
		"Filter":      filter.SystemErrorFilter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}
