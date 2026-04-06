package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/ops/generic/ratelimit/application/dto"
	ratelimitentity "gct/internal/context/ops/generic/ratelimit/domain/entity"
	ratelimitrepo "gct/internal/context/ops/generic/ratelimit/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetRateLimitQuery holds the input for getting a single rate limit.
type GetRateLimitQuery struct {
	ID ratelimitentity.RateLimitID
}

// GetRateLimitHandler handles the GetRateLimitQuery.
type GetRateLimitHandler struct {
	readRepo ratelimitrepo.RateLimitReadRepository
	logger   logger.Log
}

// NewGetRateLimitHandler creates a new GetRateLimitHandler.
func NewGetRateLimitHandler(readRepo ratelimitrepo.RateLimitReadRepository, l logger.Log) *GetRateLimitHandler {
	return &GetRateLimitHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetRateLimitQuery and returns a RateLimitView.
func (h *GetRateLimitHandler) Handle(ctx context.Context, q GetRateLimitQuery) (result *dto.RateLimitView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetRateLimitHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetRateLimit", "rate_limit")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetRateLimit", Entity: "rate_limit", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &dto.RateLimitView{
		ID:                uuid.UUID(v.ID),
		Name:              v.Name,
		Rule:              v.Rule,
		RequestsPerWindow: v.RequestsPerWindow,
		WindowDuration:    v.WindowDuration,
		Enabled:           v.Enabled,
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}, nil
}
