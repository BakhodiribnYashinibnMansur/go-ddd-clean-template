package command

import (
	"context"

	"gct/internal/ratelimit/domain"
	"gct/internal/shared/application"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// UpdateRateLimitCommand holds the input for updating a rate limit.
type UpdateRateLimitCommand struct {
	ID                uuid.UUID
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

	rl, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	rl.Update(cmd.Name, cmd.Rule, cmd.RequestsPerWindow, cmd.WindowDuration, cmd.Enabled)

	if err := h.repo.Update(ctx, rl); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateRateLimit", Entity: "rate_limit", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, rl.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateRateLimit", Entity: "rate_limit", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
