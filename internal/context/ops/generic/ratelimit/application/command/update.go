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

// UpdateRateLimitCommand holds the input for updating a rate limit.
type UpdateRateLimitCommand struct {
	ID                ratelimitentity.RateLimitID
	Name              *string
	Rule              *string
	RequestsPerWindow *int
	WindowDuration    *int
	Enabled           *bool
}

// UpdateRateLimitHandler handles the UpdateRateLimitCommand.
type UpdateRateLimitHandler struct {
	repo      ratelimitrepo.RateLimitRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewUpdateRateLimitHandler creates a new UpdateRateLimitHandler.
func NewUpdateRateLimitHandler(
	repo ratelimitrepo.RateLimitRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *UpdateRateLimitHandler {
	return &UpdateRateLimitHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle executes the UpdateRateLimitCommand.
func (h *UpdateRateLimitHandler) Handle(ctx context.Context, cmd UpdateRateLimitCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateRateLimitHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateRateLimit", "rate_limit")()

	rl, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	rl.Update(cmd.Name, cmd.Rule, cmd.RequestsPerWindow, cmd.WindowDuration, cmd.Enabled)

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Update(ctx, q, rl); err != nil {
			h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateRateLimit", Entity: "rate_limit", EntityID: cmd.ID, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, rl.Events)
}
