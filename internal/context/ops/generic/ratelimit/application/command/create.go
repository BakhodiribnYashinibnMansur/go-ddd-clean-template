package command

import (
	"context"

	ratelimitentity "gct/internal/context/ops/generic/ratelimit/domain/entity"
	ratelimitrepo "gct/internal/context/ops/generic/ratelimit/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
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
	repo     ratelimitrepo.RateLimitRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateRateLimitHandler creates a new CreateRateLimitHandler.
func NewCreateRateLimitHandler(
	repo ratelimitrepo.RateLimitRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateRateLimitHandler {
	return &CreateRateLimitHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateRateLimitCommand.
func (h *CreateRateLimitHandler) Handle(ctx context.Context, cmd CreateRateLimitCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateRateLimitHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateRateLimit", "rate_limit")()

	rl := ratelimitentity.NewRateLimit(cmd.Name, cmd.Rule, cmd.RequestsPerWindow, cmd.WindowDuration, cmd.Enabled)

	if err := h.repo.Save(ctx, rl); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateRateLimit", Entity: "rate_limit", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, rl.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateRateLimit", Entity: "rate_limit", Err: err}.KV()...)
	}

	return nil
}
