package query

import (
	"context"

	appdto "gct/internal/errorcode/application"
	"gct/internal/errorcode/domain"

	"github.com/google/uuid"
)

// GetErrorCodeQuery holds the input for getting a single error code.
type GetErrorCodeQuery struct {
	ID uuid.UUID
}

// GetErrorCodeHandler handles the GetErrorCodeQuery.
type GetErrorCodeHandler struct {
	readRepo domain.ErrorCodeReadRepository
}

// NewGetErrorCodeHandler creates a new GetErrorCodeHandler.
func NewGetErrorCodeHandler(readRepo domain.ErrorCodeReadRepository) *GetErrorCodeHandler {
	return &GetErrorCodeHandler{readRepo: readRepo}
}

// Handle executes the GetErrorCodeQuery and returns an ErrorCodeView.
func (h *GetErrorCodeHandler) Handle(ctx context.Context, q GetErrorCodeQuery) (*appdto.ErrorCodeView, error) {
	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.ErrorCodeView{
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
	}, nil
}
