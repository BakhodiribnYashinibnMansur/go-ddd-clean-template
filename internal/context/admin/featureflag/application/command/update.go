package command

import (
	"context"

	"gct/internal/context/admin/featureflag/domain"
	"gct/internal/platform/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"

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
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateFeatureFlag", "feature_flag")()

	ff, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	ff.UpdateDetails(cmd.Name, cmd.Key, cmd.Description, cmd.FlagType, cmd.DefaultValue, cmd.RolloutPercentage, cmd.IsActive)

	if err := h.repo.Update(ctx, ff); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "UpdateFeatureFlag", Entity: "feature_flag", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	ff.AddEvent(domain.NewFlagUpdated(ff.ID()))

	if err := h.eventBus.Publish(ctx, ff.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateFeatureFlag", Entity: "feature_flag", Err: err}.KV()...)
	}

	return nil
}
