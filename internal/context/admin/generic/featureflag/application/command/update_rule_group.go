package command

import (
	"context"
	"fmt"

	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"
	ffevent "gct/internal/context/admin/generic/featureflag/domain/event"
	ffrepo "gct/internal/context/admin/generic/featureflag/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// UpdateRuleGroupCommand represents a partial update to an existing rule group.
type UpdateRuleGroupCommand struct {
	ID         ffentity.RuleGroupID
	Name       *string
	Variation  *string
	Priority   *int
	Conditions *[]ConditionInput
}

// UpdateRuleGroupHandler applies modifications to an existing rule group.
type UpdateRuleGroupHandler struct {
	rgRepo   ffrepo.RuleGroupRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateRuleGroupHandler wires dependencies for rule group updates.
func NewUpdateRuleGroupHandler(
	rgRepo ffrepo.RuleGroupRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateRuleGroupHandler {
	return &UpdateRuleGroupHandler{
		rgRepo:   rgRepo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle fetches the rule group, applies updates, and persists.
func (h *UpdateRuleGroupHandler) Handle(ctx context.Context, cmd UpdateRuleGroupCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateRuleGroupHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateRuleGroup", "rule_group")()

	rg, err := h.rgRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	rg.UpdateDetails(cmd.Name, cmd.Variation, cmd.Priority)

	if cmd.Conditions != nil {
		// Validate operators.
		for _, c := range *cmd.Conditions {
			if !ffentity.IsValidOperator(c.Operator) {
				return apperrors.MapToServiceError(fmt.Errorf("%w: %s", ffentity.ErrInvalidOperator, c.Operator))
			}
		}

		// Build new conditions and reconstruct the rule group with them.
		var newConditions []ffentity.Condition
		for _, c := range *cmd.Conditions {
			newConditions = append(newConditions, ffentity.NewCondition(c.Attribute, c.Operator, c.Value))
		}

		rg = ffentity.ReconstructRuleGroup(
			rg.ID(), rg.FlagID(), rg.Name(), rg.Variation(), rg.Priority(),
			rg.CreatedAt(), rg.UpdatedAt(), newConditions,
		)
	}

	if err := h.rgRepo.Update(ctx, rg); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "UpdateRuleGroup", Entity: "rule_group", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, ffevent.NewFlagUpdated(rg.FlagID())); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateRuleGroup", Entity: "rule_group", Err: err}.KV()...)
	}

	return nil
}
