package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteCommand represents an intent to permanently remove a feature flag.
// Once deleted, any code paths checking this flag will no longer find it — callers should ensure
// application code handles missing flags gracefully (e.g., defaulting to disabled).
type DeleteCommand struct {
	ID uuid.UUID
}

// DeleteHandler performs hard deletion of a feature flag via the repository.
// Despite having an event bus dependency, no events are currently emitted on deletion.
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

// Handle deletes the feature flag identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found) to the caller.
func (h *DeleteHandler) Handle(ctx context.Context, cmd DeleteCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete feature flag: %v", err)
		return err
	}

	return nil
}
