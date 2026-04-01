package query

import (
	"context"

	"gct/internal/shared/infrastructure/pgxutil"
	appdto "gct/internal/systemerror/application"
	"gct/internal/systemerror/domain"

	"github.com/google/uuid"
)

// GetSystemErrorQuery holds the input for fetching a single system error.
type GetSystemErrorQuery struct {
	ID uuid.UUID
}

// GetSystemErrorHandler handles the GetSystemErrorQuery.
type GetSystemErrorHandler struct {
	readRepo domain.SystemErrorReadRepository
}

// NewGetSystemErrorHandler creates a new GetSystemErrorHandler.
func NewGetSystemErrorHandler(readRepo domain.SystemErrorReadRepository) *GetSystemErrorHandler {
	return &GetSystemErrorHandler{readRepo: readRepo}
}

// Handle executes the GetSystemErrorQuery and returns a SystemErrorView.
func (h *GetSystemErrorHandler) Handle(ctx context.Context, q GetSystemErrorQuery) (_ *appdto.SystemErrorView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetSystemErrorHandler.Handle")
	defer func() { end(err) }()

	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.SystemErrorView{
		ID:          view.ID,
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
