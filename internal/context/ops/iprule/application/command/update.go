package command

import (
	"context"
	"time"

	"gct/internal/context/ops/iprule/domain"
	"gct/internal/platform/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// UpdateIPRuleCommand represents a partial update to an existing IP rule identified by ID.
// Pointer fields implement patch semantics — nil means "leave unchanged," non-nil means "overwrite."
// Changing Action or IPAddress takes effect immediately for subsequent traffic evaluations.
type UpdateIPRuleCommand struct {
	ID        uuid.UUID
	IPAddress *string
	Action    *string
	Reason    *string
	ExpiresAt *time.Time
}

// UpdateIPRuleHandler applies partial modifications to an existing IP rule via fetch-then-update.
// Domain events are emitted so downstream caches or firewalls can refresh their rule sets.
type UpdateIPRuleHandler struct {
	repo     domain.IPRuleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateIPRuleHandler wires up the handler with its required dependencies.
func NewUpdateIPRuleHandler(
	repo domain.IPRuleRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateIPRuleHandler {
	return &UpdateIPRuleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle fetches the IP rule by ID, applies the patch via domain logic, and persists the result.
// Returns a repository error if the rule is not found. Event publish failures are logged but non-fatal.
func (h *UpdateIPRuleHandler) Handle(ctx context.Context, cmd UpdateIPRuleCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateIPRuleHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateIPRule", "ip_rule")()

	r, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	r.Update(cmd.IPAddress, cmd.Action, cmd.Reason, cmd.ExpiresAt)

	if err := h.repo.Update(ctx, r); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateIPRule", Entity: "ip_rule", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, r.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateIPRule", Entity: "ip_rule", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
