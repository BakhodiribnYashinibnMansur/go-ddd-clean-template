package command

import (
	"context"
	"fmt"

	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"
	ffevent "gct/internal/context/admin/generic/featureflag/domain/event"
	ffrepo "gct/internal/context/admin/generic/featureflag/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
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
	repo      ffrepo.FeatureFlagRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateHandler wires dependencies for feature flag creation.
func NewCreateHandler(
	repo ffrepo.FeatureFlagRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateHandler {
	return &CreateHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle persists a new feature flag and publishes its domain events.
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateFeatureFlag", "feature_flag")()

	ff, err := ffentity.NewFeatureFlag(cmd.Name, cmd.Key, cmd.Description, cmd.FlagType, cmd.DefaultValue, cmd.RolloutPercentage)
	if err != nil {
		return fmt.Errorf("create_feature_flag: %w", err)
	}

	if cmd.IsActive {
		ff.Activate()
	}

	ff.AddEvent(ffevent.NewFlagCreated(ff.ID()))

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Save(ctx, q, ff); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateFeatureFlag", Entity: "feature_flag", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, ff.Events)
}
