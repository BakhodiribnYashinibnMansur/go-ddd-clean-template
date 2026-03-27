package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateCommand represents a partial update to an existing feature flag.
// Nil pointer fields are left unchanged, enabling callers to toggle Enabled or adjust RolloutPercentage independently.
// Changes take effect on the next flag evaluation — there is no versioning or rollback at this level.
type UpdateCommand struct {
	ID                uuid.UUID
	Name              *string
	Description       *string
	Enabled           *bool
	RolloutPercentage *int
}

// UpdateHandler applies partial modifications to an existing feature flag using a fetch-mutate-persist pattern.
// Event publish failures are logged but do not roll back the persisted changes.
type UpdateHandler struct {
	repo     domain.FeatureFlagRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateHandler wires dependencies for feature flag updates.
func NewUpdateHandler(
	repo domain.FeatureFlagRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateHandler {
	return &UpdateHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle fetches the flag by ID, applies non-nil field updates, persists, and publishes events.
// Returns a repository error if the flag is not found or the update fails.
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) error {
	ff, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	ff.UpdateDetails(cmd.Name, cmd.Description, cmd.Enabled, cmd.RolloutPercentage)

	if err := h.repo.Update(ctx, ff); err != nil {
		h.logger.Errorf("failed to update feature flag: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, ff.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
