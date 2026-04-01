package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
)

// CreateCommand represents an intent to register a new feature flag.
type CreateCommand struct {
	Name              string
	Key               string
	Description       string
	FlagType          string
	DefaultValue      string
	RolloutPercentage int
	IsActive          bool
}

// CreateHandler orchestrates feature flag creation and emits domain events.
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
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateHandler.Handle")
	defer func() { end(err) }()

	ff := domain.NewFeatureFlag(cmd.Name, cmd.Key, cmd.Description, cmd.FlagType, cmd.DefaultValue, cmd.RolloutPercentage)

	if cmd.IsActive {
		ff.Activate()
	}

	if err := h.repo.Save(ctx, ff); err != nil {
		h.logger.Errorf("failed to save feature flag: %v", err)
		return err
	}

	ff.AddEvent(domain.NewFlagCreated(ff.ID()))

	if err := h.eventBus.Publish(ctx, ff.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
