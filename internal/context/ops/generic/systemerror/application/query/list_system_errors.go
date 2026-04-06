package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/ops/generic/systemerror/application/dto"
	syserrrepo "gct/internal/context/ops/generic/systemerror/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListSystemErrorsQuery holds the input for listing system errors with filtering.
type ListSystemErrorsQuery struct {
	Filter syserrrepo.SystemErrorFilter
}

// ListSystemErrorsResult holds the output of the list system errors query.
type ListSystemErrorsResult struct {
	Errors []*dto.SystemErrorView
	Total  int64
}

// ListSystemErrorsHandler handles the ListSystemErrorsQuery.
type ListSystemErrorsHandler struct {
	readRepo syserrrepo.SystemErrorReadRepository
	logger   logger.Log
}

// NewListSystemErrorsHandler creates a new ListSystemErrorsHandler.
func NewListSystemErrorsHandler(readRepo syserrrepo.SystemErrorReadRepository, l logger.Log) *ListSystemErrorsHandler {
	return &ListSystemErrorsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListSystemErrorsQuery and returns a list of SystemErrorView with total count.
func (h *ListSystemErrorsHandler) Handle(ctx context.Context, q ListSystemErrorsQuery) (_ *ListSystemErrorsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListSystemErrorsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListSystemErrors", "system_error")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListSystemErrors", Entity: "system_error", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*dto.SystemErrorView, len(views))
	for i, v := range views {
		result[i] = &dto.SystemErrorView{
			ID:          uuid.UUID(v.ID),
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
