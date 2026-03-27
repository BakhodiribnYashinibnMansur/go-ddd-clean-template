package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateCommand represents an intent to register a new feature flag for gradual rollout control.
// Enabled controls the global kill-switch; RolloutPercentage (0-100) determines what fraction of eligible
// users see the feature when enabled. A flag created with Enabled=false is dormant regardless of rollout.
type CreateCommand struct {
	Name              string
	Description       string
	Enabled           bool
	RolloutPercentage int
}

// CreateHandler orchestrates feature flag creation and emits domain events for downstream cache/config systems.
// Event publish failures are logged but do not roll back the persisted flag.
type CreateHandler struct {
	repo     domain.FeatureFlagRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateHandler wires dependencies for feature flag creation.
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

// Handle persists a new feature flag and publishes its domain events.
// Returns nil on success; propagates repository errors (e.g., duplicate name) to the caller.
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
