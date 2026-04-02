package command

import (
	"context"
	"fmt"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// UpdateRuleGroupCommand represents a partial update to an existing rule group.
type UpdateRuleGroupCommand struct {
	ID         uuid.UUID
	Name       *string
	Variation  *string
	Priority   *int
	Conditions *[]ConditionInput
}

// UpdateRuleGroupHandler applies modifications to an existing rule group.
type UpdateRuleGroupHandler struct {
	rgRepo   domain.RuleGroupRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateRuleGroupHandler wires dependencies for rule group updates.
func NewUpdateRuleGroupHandler(
	rgRepo domain.RuleGroupRepository,
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

	rg, err := h.rgRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	rg.UpdateDetails(cmd.Name, cmd.Variation, cmd.Priority)

	if cmd.Conditions != nil {
		// Validate operators.
		for _, c := range *cmd.Conditions {
			if !domain.IsValidOperator(c.Operator) {
				return apperrors.MapToServiceError(fmt.Errorf("%w: %s", domain.ErrInvalidOperator, c.Operator))
			}
		}

		// Build new conditions and reconstruct the rule group with them.
		var newConditions []domain.Condition
		for _, c := range *cmd.Conditions {
			newConditions = append(newConditions, domain.NewCondition(c.Attribute, c.Operator, c.Value))
		}

		rg = domain.ReconstructRuleGroup(
			rg.ID(), rg.FlagID(), rg.Name(), rg.Variation(), rg.Priority(),
			rg.CreatedAt(), rg.UpdatedAt(), newConditions,
		)
	}

	if err := h.rgRepo.Update(ctx, rg); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "UpdateRuleGroup", Entity: "rule_group", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, domain.NewFlagUpdated(rg.FlagID())); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateRuleGroup", Entity: "rule_group", Err: err}.KV()...)
	}

	return nil
}
