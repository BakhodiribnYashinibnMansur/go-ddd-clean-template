package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateCommand holds the input for updating a feature flag.
type UpdateCommand struct {
	ID                uuid.UUID
	Name              *string
	Description       *string
	Enabled           *bool
	RolloutPercentage *int
}

// UpdateHandler handles the UpdateCommand.
type UpdateHandler struct {
	repo     domain.FeatureFlagRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateHandler creates a new UpdateHandler.
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

// Handle executes the UpdateCommand.
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
