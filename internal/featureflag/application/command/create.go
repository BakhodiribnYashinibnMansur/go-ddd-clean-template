package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateCommand holds the input for creating a new feature flag.
type CreateCommand struct {
	Name              string
	Description       string
	Enabled           bool
	RolloutPercentage int
}

// CreateHandler handles the CreateCommand.
type CreateHandler struct {
	repo     domain.FeatureFlagRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateHandler creates a new CreateHandler.
func NewCreateHandler(
	repo domain.FeatureFlagRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateHandler {
	return &CreateHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateCommand.
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) error {
	ff := domain.NewFeatureFlag(cmd.Name, cmd.Description, cmd.Enabled, cmd.RolloutPercentage)

	if err := h.repo.Save(ctx, ff); err != nil {
		h.logger.Errorf("failed to save feature flag: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, ff.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
