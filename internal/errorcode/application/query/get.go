package query

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"

	appdto "gct/internal/errorcode/application"
	"gct/internal/errorcode/domain"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetErrorCodeQuery holds the input for getting a single error code.
type GetErrorCodeQuery struct {
	ID uuid.UUID
}

// GetErrorCodeHandler handles the GetErrorCodeQuery.
type GetErrorCodeHandler struct {
	readRepo domain.ErrorCodeReadRepository
	logger   logger.Log
}

// NewGetErrorCodeHandler creates a new GetErrorCodeHandler.
func NewGetErrorCodeHandler(readRepo domain.ErrorCodeReadRepository, l logger.Log) *GetErrorCodeHandler {
	return &GetErrorCodeHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetErrorCodeQuery and returns an ErrorCodeView.
func (h *GetErrorCodeHandler) Handle(ctx context.Context, q GetErrorCodeQuery) (result *appdto.ErrorCodeView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetErrorCodeHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetErrorCode", "error_code")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetErrorCode", Entity: "error_code", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &appdto.ErrorCodeView{
		ID:         v.ID,
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
	}, nil
}
