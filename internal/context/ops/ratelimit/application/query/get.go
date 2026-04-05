package query

import (
	"context"

	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"

	appdto "gct/internal/context/ops/ratelimit/application"
	"gct/internal/context/ops/ratelimit/domain"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetRateLimitQuery holds the input for getting a single rate limit.
type GetRateLimitQuery struct {
	ID uuid.UUID
}

// GetRateLimitHandler handles the GetRateLimitQuery.
type GetRateLimitHandler struct {
	readRepo domain.RateLimitReadRepository
	logger   logger.Log
}

// NewGetRateLimitHandler creates a new GetRateLimitHandler.
func NewGetRateLimitHandler(readRepo domain.RateLimitReadRepository, l logger.Log) *GetRateLimitHandler {
	return &GetRateLimitHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetRateLimitQuery and returns a RateLimitView.
func (h *GetRateLimitHandler) Handle(ctx context.Context, q GetRateLimitQuery) (result *appdto.RateLimitView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetRateLimitHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetRateLimit", "rate_limit")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetRateLimit", Entity: "rate_limit", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
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
