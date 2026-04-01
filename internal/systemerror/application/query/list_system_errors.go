package query

import (
	"context"

	"gct/internal/shared/infrastructure/pgxutil"
	appdto "gct/internal/systemerror/application"
	"gct/internal/systemerror/domain"
)

// ListSystemErrorsQuery holds the input for listing system errors with filtering.
type ListSystemErrorsQuery struct {
	Filter domain.SystemErrorFilter
}

// ListSystemErrorsResult holds the output of the list system errors query.
type ListSystemErrorsResult struct {
	Errors []*appdto.SystemErrorView
	Total  int64
}

// ListSystemErrorsHandler handles the ListSystemErrorsQuery.
type ListSystemErrorsHandler struct {
	readRepo domain.SystemErrorReadRepository
}

// NewListSystemErrorsHandler creates a new ListSystemErrorsHandler.
func NewListSystemErrorsHandler(readRepo domain.SystemErrorReadRepository) *ListSystemErrorsHandler {
	return &ListSystemErrorsHandler{readRepo: readRepo}
}

// Handle executes the ListSystemErrorsQuery and returns a list of SystemErrorView with total count.
func (h *ListSystemErrorsHandler) Handle(ctx context.Context, q ListSystemErrorsQuery) (_ *ListSystemErrorsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListSystemErrorsHandler.Handle")
	defer func() { end(err) }()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.SystemErrorView, len(views))
	for i, v := range views {
		result[i] = &appdto.SystemErrorView{
			ID:          v.ID,
			Code:        v.Code,
			Message:     v.Message,
			StackTrace:  v.StackTrace,
			Metadata:    v.Metadata,
			Severity:    v.Severity,
			ServiceName: v.ServiceName,
			RequestID:   v.RequestID,
			UserID:      v.UserID,
			IPAddress:   v.IPAddress,
			Path:        v.Path,
			Method:      v.Method,
			IsResolved:  v.IsResolved,
			ResolvedAt:  v.ResolvedAt,
			ResolvedBy:  v.ResolvedBy,
			CreatedAt:   v.CreatedAt,
		}
	}

	return &ListSystemErrorsResult{
		Errors: result,
		Total:  total,
	}, nil
}
