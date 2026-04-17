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

// CreateRateLimitCommand holds the input for creating a new rate limit.
type CreateRateLimitCommand struct {
	Name              string
	Rule              string
	RequestsPerWindow int
	WindowDuration    int
	Enabled           bool
}

// CreateRateLimitHandler handles the CreateRateLimitCommand.
type CreateRateLimitHandler struct {
	repo      ratelimitrepo.RateLimitRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateRateLimitHandler creates a new CreateRateLimitHandler.
func NewCreateRateLimitHandler(
	repo ratelimitrepo.RateLimitRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateRateLimitHandler {
	return &CreateRateLimitHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle executes the CreateRateLimitCommand.
func (h *CreateRateLimitHandler) Handle(ctx context.Context, cmd CreateRateLimitCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateRateLimitHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateRateLimit", "rate_limit")()

	rl := ratelimitentity.NewRateLimit(cmd.Name, cmd.Rule, cmd.RequestsPerWindow, cmd.WindowDuration, cmd.Enabled)

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Save(ctx, q, rl); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateRateLimit", Entity: "rate_limit", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, rl.Events)
}
