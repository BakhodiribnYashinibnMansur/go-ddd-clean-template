package command

import (
	"context"
	"fmt"

	"gct/internal/context/admin/generic/featureflag/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ConditionInput is the input DTO for creating a condition.
type ConditionInput struct {
	Attribute string
	Operator  string
	Value     string
}

// CreateRuleGroupCommand represents an intent to add a rule group to a feature flag.
type CreateRuleGroupCommand struct {
	FlagID     domain.FeatureFlagID
	Name       string
	Variation  string
	Priority   int
	Conditions []ConditionInput
}

// CreateRuleGroupHandler orchestrates rule group creation.
type CreateRuleGroupHandler struct {
	flagRepo domain.FeatureFlagRepository
	rgRepo   domain.RuleGroupRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateRuleGroupHandler wires dependencies for rule group creation.
func NewCreateRuleGroupHandler(
	flagRepo domain.FeatureFlagRepository,
	rgRepo domain.RuleGroupRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateRuleGroupHandler {
	return &CreateRuleGroupHandler{
		flagRepo: flagRepo,
		rgRepo:   rgRepo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle creates a new rule group with conditions for the given flag.
func (h *CreateRuleGroupHandler) Handle(ctx context.Context, cmd CreateRuleGroupCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateRuleGroupHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateRuleGroup", "rule_group")()

	// Verify the flag exists.
	if _, err := h.flagRepo.FindByID(ctx, cmd.FlagID); err != nil {
		return apperrors.MapToServiceError(err)
	}

	// Validate operators.
	for _, c := range cmd.Conditions {
		if !domain.IsValidOperator(c.Operator) {
			return apperrors.MapToServiceError(fmt.Errorf("%w: %s", domain.ErrInvalidOperator, c.Operator))
		}
	}

	rg := domain.NewRuleGroup(cmd.FlagID.UUID(), cmd.Name, cmd.Variation, cmd.Priority)

	for _, c := range cmd.Conditions {
		rg.AddCondition(domain.NewCondition(c.Attribute, c.Operator, c.Value))
	}

	if err := h.rgRepo.Save(ctx, rg); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateRuleGroup", Entity: "rule_group", EntityID: cmd.FlagID.String(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, domain.NewFlagUpdated(cmd.FlagID.UUID())); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateRuleGroup", Entity: "rule_group", Err: err}.KV()...)
	}

	return nil
}
