package command

import (
	"context"

	"gct/internal/ratelimit/domain"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

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

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete rate limit: %v", err)
		return err
	}
	return nil
}
