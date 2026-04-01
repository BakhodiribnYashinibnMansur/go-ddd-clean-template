package query

import (
	"context"

	appdto "gct/internal/audit/application"
	"gct/internal/audit/domain"
	"gct/internal/shared/infrastructure/pgxutil"
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
}

// NewListEndpointHistoryHandler creates a new ListEndpointHistoryHandler.
func NewListEndpointHistoryHandler(readRepo domain.AuditReadRepository) *ListEndpointHistoryHandler {
	return &ListEndpointHistoryHandler{readRepo: readRepo}
}

// Handle executes the ListEndpointHistoryQuery and returns a list of EndpointHistoryView with total count.
func (h *ListEndpointHistoryHandler) Handle(ctx context.Context, q ListEndpointHistoryQuery) (_ *ListEndpointHistoryResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListEndpointHistoryHandler.Handle")
	defer func() { end(err) }()

	views, total, err := h.readRepo.ListEndpointHistory(ctx, q.Filter)
	if err != nil {
		return nil, err
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
