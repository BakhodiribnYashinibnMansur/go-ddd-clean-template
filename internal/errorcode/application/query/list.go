package query

import (
	"context"

	appdto "gct/internal/errorcode/application"
	"gct/internal/errorcode/domain"
)

// ListErrorCodesQuery holds the input for listing error codes with filtering.
type ListErrorCodesQuery struct {
	Filter domain.ErrorCodeFilter
}

// ListErrorCodesResult holds the output of the list error codes query.
type ListErrorCodesResult struct {
	ErrorCodes []*appdto.ErrorCodeView
	Total      int64
}

// ListErrorCodesHandler handles the ListErrorCodesQuery.
type ListErrorCodesHandler struct {
	readRepo domain.ErrorCodeReadRepository
}

// NewListErrorCodesHandler creates a new ListErrorCodesHandler.
func NewListErrorCodesHandler(readRepo domain.ErrorCodeReadRepository) *ListErrorCodesHandler {
	return &ListErrorCodesHandler{readRepo: readRepo}
}

// Handle executes the ListErrorCodesQuery and returns a list of ErrorCodeView with total count.
func (h *ListErrorCodesHandler) Handle(ctx context.Context, q ListErrorCodesQuery) (*ListErrorCodesResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.ErrorCodeView, len(views))
	for i, v := range views {
		result[i] = &appdto.ErrorCodeView{
			ID:         v.ID,
			Code:       v.Code,
			Message:    v.Message,
			HTTPStatus: v.HTTPStatus,
			Category:   v.Category,
			Severity:   v.Severity,
			Retryable:  v.Retryable,
			RetryAfter: v.RetryAfter,
			Suggestion: v.Suggestion,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
		}
	}

	return &ListErrorCodesResult{
		ErrorCodes: result,
		Total:      total,
	}, nil
}
