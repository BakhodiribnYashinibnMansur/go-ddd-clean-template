package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/ops/generic/systemerror/application/dto"
	syserrentity "gct/internal/context/ops/generic/systemerror/domain/entity"
	syserrrepo "gct/internal/context/ops/generic/systemerror/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetSystemErrorQuery holds the input for fetching a single system error.
type GetSystemErrorQuery struct {
	ID syserrentity.SystemErrorID
}

// GetSystemErrorHandler handles the GetSystemErrorQuery.
type GetSystemErrorHandler struct {
	readRepo syserrrepo.SystemErrorReadRepository
	logger   logger.Log
}

// NewGetSystemErrorHandler creates a new GetSystemErrorHandler.
func NewGetSystemErrorHandler(readRepo syserrrepo.SystemErrorReadRepository, l logger.Log) *GetSystemErrorHandler {
	return &GetSystemErrorHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetSystemErrorQuery and returns a SystemErrorView.
func (h *GetSystemErrorHandler) Handle(ctx context.Context, q GetSystemErrorQuery) (_ *dto.SystemErrorView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetSystemErrorHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetSystemError", "system_error")()

	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetSystemError", Entity: "system_error", EntityID: q.ID.String(), Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &dto.SystemErrorView{
		ID:          uuid.UUID(view.ID),
		Code:        view.Code,
		Message:     view.Message,
		StackTrace:  view.StackTrace,
		Metadata:    view.Metadata,
		Severity:    view.Severity,
		ServiceName: view.ServiceName,
		RequestID:   view.RequestID,
		UserID:      view.UserID,
		IPAddress:   view.IPAddress,
		Path:        view.Path,
		Method:      view.Method,
		IsResolved:  view.IsResolved,
		ResolvedAt:  view.ResolvedAt,
		ResolvedBy:  view.ResolvedBy,
		CreatedAt:   view.CreatedAt,
	}, nil
}
