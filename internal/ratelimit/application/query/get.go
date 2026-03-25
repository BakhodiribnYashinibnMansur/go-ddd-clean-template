package query

import (
	"context"

	appdto "gct/internal/ratelimit/application"
	"gct/internal/ratelimit/domain"

	"github.com/google/uuid"
)

// GetRateLimitQuery holds the input for getting a single rate limit.
type GetRateLimitQuery struct {
	ID uuid.UUID
}

// GetRateLimitHandler handles the GetRateLimitQuery.
type GetRateLimitHandler struct {
	readRepo domain.RateLimitReadRepository
}

// NewGetRateLimitHandler creates a new GetRateLimitHandler.
func NewGetRateLimitHandler(readRepo domain.RateLimitReadRepository) *GetRateLimitHandler {
	return &GetRateLimitHandler{readRepo: readRepo}
}

// Handle executes the GetRateLimitQuery and returns a RateLimitView.
func (h *GetRateLimitHandler) Handle(ctx context.Context, q GetRateLimitQuery) (*appdto.RateLimitView, error) {
	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.RateLimitView{
		ID:                v.ID,
		Name:              v.Name,
		Rule:              v.Rule,
		RequestsPerWindow: v.RequestsPerWindow,
		WindowDuration:    v.WindowDuration,
		Enabled:           v.Enabled,
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}, nil
}
