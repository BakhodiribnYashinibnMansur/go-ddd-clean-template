package command

import (
	"context"
	"time"

	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"
	iprulerepo "gct/internal/context/ops/supporting/iprule/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// CreateIPRuleCommand represents an intent to add a new IP-based access control rule.
// Action determines the enforcement behavior (e.g., "allow" or "block") for matching traffic.
// ExpiresAt is optional — nil creates a permanent rule; otherwise the rule auto-expires at the given time.
type CreateIPRuleCommand struct {
	IPAddress string
	Action    string
	Reason    string
	ExpiresAt *time.Time
}

// CreateIPRuleHandler persists a new IP rule and emits domain events for downstream enforcement.
// Callers are responsible for validating IP format and ensuring no conflicting rule already exists.
type CreateIPRuleHandler struct {
	repo     iprulerepo.IPRuleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateIPRuleHandler wires up the handler with its required dependencies.
func NewCreateIPRuleHandler(
	repo iprulerepo.IPRuleRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateIPRuleHandler {
	return &CreateIPRuleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle creates the IP rule domain entity, persists it, and publishes domain events (e.g., IPRuleCreated).
// Event publish failures are logged but do not fail the operation — the rule is already saved.
func (h *CreateIPRuleHandler) Handle(ctx context.Context, cmd CreateIPRuleCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateIPRuleHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateIPRule", "ip_rule")()

	r := ipruleentity.NewIPRule(cmd.IPAddress, cmd.Action, cmd.Reason, cmd.ExpiresAt)

	if err := h.repo.Save(ctx, r); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateIPRule", Entity: "ip_rule", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, r.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateIPRule", Entity: "ip_rule", Err: err}.KV()...)
	}

	return nil
}
