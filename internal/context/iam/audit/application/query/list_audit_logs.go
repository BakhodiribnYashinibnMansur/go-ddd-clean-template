package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	appdto "gct/internal/context/iam/audit/application"
	"gct/internal/context/iam/audit/domain"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ListAuditLogsQuery holds the input for listing audit logs with filtering.
type ListAuditLogsQuery struct {
	Filter domain.AuditLogFilter
}

// ListAuditLogsResult holds the output of the list audit logs query.
type ListAuditLogsResult struct {
	AuditLogs []*appdto.AuditLogView
	Total     int64
}

// ListAuditLogsHandler handles the ListAuditLogsQuery.
type ListAuditLogsHandler struct {
	readRepo domain.AuditReadRepository
	logger   logger.Log
}

// NewListAuditLogsHandler creates a new ListAuditLogsHandler.
func NewListAuditLogsHandler(readRepo domain.AuditReadRepository, l logger.Log) *ListAuditLogsHandler {
	return &ListAuditLogsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListAuditLogsQuery and returns a list of AuditLogView with total count.
func (h *ListAuditLogsHandler) Handle(ctx context.Context, q ListAuditLogsQuery) (_ *ListAuditLogsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListAuditLogsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListAuditLogs", "audit_log")()

	views, total, err := h.readRepo.ListAuditLogs(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListAuditLogs", Entity: "audit_log", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*appdto.AuditLogView, len(views))
	for i, v := range views {
		result[i] = &appdto.AuditLogView{
			ID:           v.ID,
			UserID:       v.UserID,
			SessionID:    v.SessionID,
			Action:       string(v.Action),
			ResourceType: v.ResourceType,
			ResourceID:   v.ResourceID,
			Platform:     v.Platform,
			IPAddress:    v.IPAddress,
			UserAgent:    v.UserAgent,
			Permission:   v.Permission,
			PolicyID:     v.PolicyID,
			Decision:     v.Decision,
			Success:      v.Success,
			ErrorMessage: v.ErrorMessage,
			Metadata:     v.Metadata,
			CreatedAt:    v.CreatedAt,
		}
	}

	return &ListAuditLogsResult{
		AuditLogs: result,
		Total:     total,
	}, nil
}
