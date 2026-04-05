package command

import (
	"context"

	"gct/internal/context/ops/ratelimit/domain"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// DeleteRateLimitCommand holds the input for deleting a rate limit.
type DeleteRateLimitCommand struct {
	ID uuid.UUID
}

// DeleteRateLimitHandler handles the DeleteRateLimitCommand.
type DeleteRateLimitHandler struct {
	repo   domain.RateLimitRepository
	logger logger.Log
}

// NewDeleteRateLimitHandler creates a new DeleteRateLimitHandler.
func NewDeleteRateLimitHandler(
	repo domain.RateLimitRepository,
	logger logger.Log,
) *DeleteRateLimitHandler {
	return &DeleteRateLimitHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeleteRateLimitCommand.
func (h *DeleteRateLimitHandler) Handle(ctx context.Context, cmd DeleteRateLimitCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteRateLimitHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteRateLimit", "rate_limit")()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteRateLimit", Entity: "rate_limit", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}
	return nil
}
