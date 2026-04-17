package command

import (
	"context"

	ratelimitentity "gct/internal/context/ops/generic/ratelimit/domain/entity"
	ratelimitrepo "gct/internal/context/ops/generic/ratelimit/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// DeleteRateLimitCommand holds the input for deleting a rate limit.
type DeleteRateLimitCommand struct {
	ID ratelimitentity.RateLimitID
}

// DeleteRateLimitHandler handles the DeleteRateLimitCommand.
type DeleteRateLimitHandler struct {
	repo      ratelimitrepo.RateLimitRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewDeleteRateLimitHandler creates a new DeleteRateLimitHandler.
func NewDeleteRateLimitHandler(
	repo ratelimitrepo.RateLimitRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *DeleteRateLimitHandler {
	return &DeleteRateLimitHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle executes the DeleteRateLimitCommand.
func (h *DeleteRateLimitHandler) Handle(ctx context.Context, cmd DeleteRateLimitCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteRateLimitHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteRateLimit", "rate_limit")()

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Delete(ctx, q, cmd.ID); err != nil {
			h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteRateLimit", Entity: "rate_limit", EntityID: cmd.ID, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, func() []shareddomain.DomainEvent { return nil })
}
