package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/admin/supporting/errorcode/application/dto"
	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"
	errcoderepo "gct/internal/context/admin/supporting/errorcode/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetErrorCodeQuery holds the input for getting a single error code.
type GetErrorCodeQuery struct {
	ID errcodeentity.ErrorCodeID
}

// GetErrorCodeHandler handles the GetErrorCodeQuery.
type GetErrorCodeHandler struct {
	readRepo errcoderepo.ErrorCodeReadRepository
	logger   logger.Log
}

// NewGetErrorCodeHandler creates a new GetErrorCodeHandler.
func NewGetErrorCodeHandler(readRepo errcoderepo.ErrorCodeReadRepository, l logger.Log) *GetErrorCodeHandler {
	return &GetErrorCodeHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetErrorCodeQuery and returns an ErrorCodeView.
func (h *GetErrorCodeHandler) Handle(ctx context.Context, q GetErrorCodeQuery) (result *dto.ErrorCodeView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetErrorCodeHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetErrorCode", "error_code")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetErrorCode", Entity: "error_code", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &dto.ErrorCodeView{
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
	}, nil
}
