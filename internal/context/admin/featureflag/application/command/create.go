package command

import (
	"context"
	"fmt"

	"gct/internal/context/admin/featureflag/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
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
	defer logger.SlowOp(h.logger, ctx, "CreateFeatureFlag", "feature_flag")()

	ff, err := domain.NewFeatureFlag(cmd.Name, cmd.Key, cmd.Description, cmd.FlagType, cmd.DefaultValue, cmd.RolloutPercentage)
	if err != nil {
		return fmt.Errorf("create_feature_flag: %w", err)
	}

	if cmd.IsActive {
		ff.Activate()
	}

	if err := h.repo.Save(ctx, ff); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateFeatureFlag", Entity: "feature_flag", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	ff.AddEvent(domain.NewFlagCreated(ff.ID()))

	if err := h.eventBus.Publish(ctx, ff.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateFeatureFlag", Entity: "feature_flag", Err: err}.KV()...)
	}

	return nil
}
