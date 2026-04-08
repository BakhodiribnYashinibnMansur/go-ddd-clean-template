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

// ConditionInput is the input DTO for creating a condition.
type ConditionInput struct {
	Attribute string
	Operator  string
	Value     string
}

// CreateRuleGroupCommand represents an intent to add a rule group to a feature flag.
type CreateRuleGroupCommand struct {
	FlagID     ffentity.FeatureFlagID
	Name       string
	Variation  string
	Priority   int
	Conditions []ConditionInput
}

// CreateRuleGroupHandler orchestrates rule group creation.
type CreateRuleGroupHandler struct {
	flagRepo  ffrepo.FeatureFlagRepository
	rgRepo    ffrepo.RuleGroupRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateRuleGroupHandler wires dependencies for rule group creation.
func NewCreateRuleGroupHandler(
	flagRepo ffrepo.FeatureFlagRepository,
	rgRepo ffrepo.RuleGroupRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateRuleGroupHandler {
	return &CreateRuleGroupHandler{
		flagRepo:  flagRepo,
		rgRepo:    rgRepo,
		committer: committer,
		logger:    logger,
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
		if !ffentity.IsValidOperator(c.Operator) {
			return apperrors.MapToServiceError(fmt.Errorf("%w: %s", ffentity.ErrInvalidOperator, c.Operator))
		}
	}

	rg := ffentity.NewRuleGroup(cmd.FlagID.UUID(), cmd.Name, cmd.Variation, cmd.Priority)

	for _, c := range cmd.Conditions {
		rg.AddCondition(ffentity.NewCondition(c.Attribute, c.Operator, c.Value))
	}

	event := ffevent.NewFlagUpdated(cmd.FlagID.UUID())

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.rgRepo.Save(ctx, rg); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateRuleGroup", Entity: "rule_group", EntityID: cmd.FlagID.String(), Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, func() []shareddomain.DomainEvent { return []shareddomain.DomainEvent{event} })
}
