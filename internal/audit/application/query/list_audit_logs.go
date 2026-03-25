package query

import (
	"context"

	appdto "gct/internal/audit/application"
	"gct/internal/audit/domain"
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
}

// NewListAuditLogsHandler creates a new ListAuditLogsHandler.
func NewListAuditLogsHandler(readRepo domain.AuditReadRepository) *ListAuditLogsHandler {
	return &ListAuditLogsHandler{readRepo: readRepo}
}

// Handle executes the ListAuditLogsQuery and returns a list of AuditLogView with total count.
func (h *ListAuditLogsHandler) Handle(ctx context.Context, q ListAuditLogsQuery) (*ListAuditLogsResult, error) {
	views, total, err := h.readRepo.ListAuditLogs(ctx, q.Filter)
	if err != nil {
		return nil, err
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
