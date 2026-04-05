package command

import (
	"context"

	"gct/internal/context/ops/generic/ratelimit/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// UpdateRateLimitCommand holds the input for updating a rate limit.
type UpdateRateLimitCommand struct {
	ID                domain.RateLimitID
	Name              *string
	Rule              *string
	RequestsPerWindow *int
	WindowDuration    *int
	Enabled           *bool
}

// UpdateRateLimitHandler handles the UpdateRateLimitCommand.
type UpdateRateLimitHandler struct {
	repo     domain.RateLimitRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateRateLimitHandler creates a new UpdateRateLimitHandler.
func NewUpdateRateLimitHandler(
	repo domain.RateLimitRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateRateLimitHandler {
	return &UpdateRateLimitHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateRateLimitCommand.
func (h *UpdateRateLimitHandler) Handle(ctx context.Context, cmd UpdateRateLimitCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateRateLimitHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateRateLimit", "rate_limit")()

	rl, err := h.repo.FindByID(ctx, cmd.ID.UUID())
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	rl.Update(cmd.Name, cmd.Rule, cmd.RequestsPerWindow, cmd.WindowDuration, cmd.Enabled)

	if err := h.repo.Update(ctx, rl); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateRateLimit", Entity: "rate_limit", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, rl.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateRateLimit", Entity: "rate_limit", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
	}

	return nil
}
