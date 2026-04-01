package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// DeleteCommand represents an intent to permanently remove a feature flag.
type DeleteCommand struct {
	ID uuid.UUID
}

// DeleteHandler performs hard deletion of a feature flag via the repository.
type DeleteHandler struct {
	repo     domain.FeatureFlagRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteHandler wires dependencies for feature flag deletion.
func NewDeleteHandler(
	repo domain.FeatureFlagRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *DeleteHandler {
	return &DeleteHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle deletes the feature flag and publishes a FlagDeleted event.
func (h *DeleteHandler) Handle(ctx context.Context, cmd DeleteCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteHandler.Handle")
	defer func() { end(err) }()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete feature flag: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, domain.NewFlagDeleted(cmd.ID)); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
