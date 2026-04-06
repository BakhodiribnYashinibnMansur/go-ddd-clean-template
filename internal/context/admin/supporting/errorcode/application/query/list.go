package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/admin/supporting/errorcode/application/dto"
	errcoderepo "gct/internal/context/admin/supporting/errorcode/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListErrorCodesQuery holds the input for listing error codes with filtering.
type ListErrorCodesQuery struct {
	Filter errcoderepo.ErrorCodeFilter
}

// ListErrorCodesResult holds the output of the list error codes query.
type ListErrorCodesResult struct {
	ErrorCodes []*dto.ErrorCodeView
	Total      int64
}

// ListErrorCodesHandler handles the ListErrorCodesQuery.
type ListErrorCodesHandler struct {
	readRepo errcoderepo.ErrorCodeReadRepository
	logger   logger.Log
}

// NewListErrorCodesHandler creates a new ListErrorCodesHandler.
func NewListErrorCodesHandler(readRepo errcoderepo.ErrorCodeReadRepository, l logger.Log) *ListErrorCodesHandler {
	return &ListErrorCodesHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListErrorCodesQuery and returns a list of ErrorCodeView with total count.
func (h *ListErrorCodesHandler) Handle(ctx context.Context, q ListErrorCodesQuery) (_ *ListErrorCodesResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListErrorCodesHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListErrorCodes", "error_code")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListErrorCodes", Entity: "error_code", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*dto.ErrorCodeView, len(views))
	for i, v := range views {
		result[i] = &dto.ErrorCodeView{
			ID:         uuid.UUID(v.ID),
			Code:       v.Code,
			Message:    v.Message,
			MessageUz:  v.MessageUz,
			MessageRu:  v.MessageRu,
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
