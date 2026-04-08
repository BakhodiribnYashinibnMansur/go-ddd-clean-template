package command

import (
	"context"

	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"
	ffevent "gct/internal/context/admin/generic/featureflag/domain/event"
	ffrepo "gct/internal/context/admin/generic/featureflag/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// UpdateCommand represents a partial update to an existing feature flag.
type UpdateCommand struct {
	ID                ffentity.FeatureFlagID
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
	repo      ffrepo.FeatureFlagRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewUpdateHandler wires dependencies for feature flag updates.
func NewUpdateHandler(
	repo ffrepo.FeatureFlagRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *UpdateHandler {
	return &UpdateHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
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

	if err := ff.UpdateDetails(cmd.Name, cmd.Key, cmd.Description, cmd.FlagType, cmd.DefaultValue, cmd.RolloutPercentage, cmd.IsActive); err != nil {
		return err
	}

	ff.AddEvent(ffevent.NewFlagUpdated(ff.ID()))

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Update(ctx, ff); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "UpdateFeatureFlag", Entity: "feature_flag", EntityID: cmd.ID, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, ff.Events)
}
