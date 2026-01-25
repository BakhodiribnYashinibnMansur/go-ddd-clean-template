package admin

import (
	"gct/pkg/httpx"
	"gct/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) AuditLogs(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.AuditLogsFilter{
		Pagination: pagination,
	}

	if userID := httpx.GetNullStringQuery(ctx, "user_id"); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			filter.UserID = &uid
		}
	}
	if action := httpx.GetNullStringQuery(ctx, "action"); action != "" {
		a := domain.AuditActionType(action)
		filter.Action = &a
	}
	if resourceType := httpx.GetNullStringQuery(ctx, "resource_type"); resourceType != "" {
		filter.ResourceType = &resourceType
	}
	if successStr := httpx.GetNullStringQuery(ctx, "success"); successStr != "" {
		s := successStr == "true"
		filter.Success = &s
	}

	logs, count, err := h.uc.Audit.Log.Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch audit logs", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "audit/logs.html", "Audit Logs", "audit_logs", map[string]any{
		"Logs":        logs,
		"Pagination":  pagination,
		"Filter":      filter.AuditLogFilter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}
