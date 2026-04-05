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

// DeleteRuleGroupCommand represents an intent to remove a rule group.
type DeleteRuleGroupCommand struct {
	ID uuid.UUID
}

// DeleteRuleGroupHandler performs deletion of a rule group.
type DeleteRuleGroupHandler struct {
	rgRepo   domain.RuleGroupRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteRuleGroupHandler wires dependencies for rule group deletion.
func NewDeleteRuleGroupHandler(
	rgRepo domain.RuleGroupRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *DeleteRuleGroupHandler {
	return &DeleteRuleGroupHandler{
		rgRepo:   rgRepo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle deletes the rule group and publishes a FlagUpdated event for the parent flag.
func (h *DeleteRuleGroupHandler) Handle(ctx context.Context, cmd DeleteRuleGroupCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteRuleGroupHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteRuleGroup", "rule_group")()

	// Find the rule group to get its flagID before deletion.
	rg, err := h.rgRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	flagID := rg.FlagID()

	if err := h.rgRepo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteRuleGroup", Entity: "rule_group", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, domain.NewFlagUpdated(flagID)); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "DeleteRuleGroup", Entity: "rule_group", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
