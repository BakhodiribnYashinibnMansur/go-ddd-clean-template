package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	appdto "gct/internal/context/iam/audit/application"
	"gct/internal/context/iam/audit/domain"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ListEndpointHistoryQuery holds the input for listing endpoint history with filtering.
type ListEndpointHistoryQuery struct {
	Filter domain.EndpointHistoryFilter
}

// ListEndpointHistoryResult holds the output of the list endpoint history query.
type ListEndpointHistoryResult struct {
	Entries []*appdto.EndpointHistoryView
	Total   int64
}

// ListEndpointHistoryHandler handles the ListEndpointHistoryQuery.
type ListEndpointHistoryHandler struct {
	readRepo domain.AuditReadRepository
	logger   logger.Log
}

// NewListEndpointHistoryHandler creates a new ListEndpointHistoryHandler.
func NewListEndpointHistoryHandler(readRepo domain.AuditReadRepository, l logger.Log) *ListEndpointHistoryHandler {
	return &ListEndpointHistoryHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListEndpointHistoryQuery and returns a list of EndpointHistoryView with total count.
func (h *ListEndpointHistoryHandler) Handle(ctx context.Context, q ListEndpointHistoryQuery) (_ *ListEndpointHistoryResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListEndpointHistoryHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListEndpointHistory", "endpoint_history")()

	views, total, err := h.readRepo.ListEndpointHistory(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListEndpointHistory", Entity: "audit_log", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*appdto.EndpointHistoryView, len(views))
	for i, v := range views {
		result[i] = &appdto.EndpointHistoryView{
			ID:         v.ID,
			UserID:     v.UserID,
			Endpoint:   v.Endpoint,
			Method:     v.Method,
			StatusCode: v.StatusCode,
			Latency:    v.Latency,
			IPAddress:  v.IPAddress,
			UserAgent:  v.UserAgent,
			CreatedAt:  v.CreatedAt,
		}
	}

	return &ListEndpointHistoryResult{
		Entries: result,
		Total:   total,
	}, nil
}
