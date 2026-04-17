package command

import (
	"context"

	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"
	ffevent "gct/internal/context/admin/generic/featureflag/domain/event"
	ffrepo "gct/internal/context/admin/generic/featureflag/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// DeleteRuleGroupCommand represents an intent to remove a rule group.
type DeleteRuleGroupCommand struct {
	ID ffentity.RuleGroupID
}

// DeleteRuleGroupHandler performs deletion of a rule group.
type DeleteRuleGroupHandler struct {
	rgRepo    ffrepo.RuleGroupRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewDeleteRuleGroupHandler wires dependencies for rule group deletion.
func NewDeleteRuleGroupHandler(
	rgRepo ffrepo.RuleGroupRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *DeleteRuleGroupHandler {
	return &DeleteRuleGroupHandler{
		rgRepo:    rgRepo,
		committer: committer,
		logger:    logger,
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
	event := ffevent.NewFlagUpdated(flagID)

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.rgRepo.Delete(ctx, q, cmd.ID); err != nil {
			h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteRuleGroup", Entity: "rule_group", EntityID: cmd.ID, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, func() []shareddomain.DomainEvent { return []shareddomain.DomainEvent{event} })
}
