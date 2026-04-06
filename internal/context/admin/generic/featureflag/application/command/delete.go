package command

import (
	"context"

	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"
	ffevent "gct/internal/context/admin/generic/featureflag/domain/event"
	ffrepo "gct/internal/context/admin/generic/featureflag/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteCommand represents an intent to permanently remove a feature flag.
type DeleteCommand struct {
	ID ffentity.FeatureFlagID
}

// DeleteHandler performs hard deletion of a feature flag via the repository.
type DeleteHandler struct {
	repo     ffrepo.FeatureFlagRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteHandler wires dependencies for feature flag deletion.
func NewDeleteHandler(
	repo ffrepo.FeatureFlagRepository,
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
	defer logger.SlowOp(h.logger, ctx, "DeleteFeatureFlag", "feature_flag")()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteFeatureFlag", Entity: "feature_flag", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, ffevent.NewFlagDeleted(cmd.ID.UUID())); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "DeleteFeatureFlag", Entity: "feature_flag", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
