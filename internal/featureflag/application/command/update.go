package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateCommand represents a partial update to an existing feature flag.
type UpdateCommand struct {
	ID                uuid.UUID
	Name              *string
	Key               *string
	Description       *string
	FlagType          *string
	DefaultValue      *string
	RolloutPercentage *int
	IsActive          *bool
}

// UpdateHandler applies partial modifications to an existing feature flag.
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
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) error {
	ff, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	ff.UpdateDetails(cmd.Name, cmd.Key, cmd.Description, cmd.FlagType, cmd.DefaultValue, cmd.RolloutPercentage, cmd.IsActive)

	if err := h.repo.Update(ctx, ff); err != nil {
		h.logger.Errorf("failed to update feature flag: %v", err)
		return err
	}

	ff.AddEvent(domain.NewFlagUpdated(ff.ID()))

	if err := h.eventBus.Publish(ctx, ff.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
