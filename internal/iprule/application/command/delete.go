package command

import (
	"context"

	"gct/internal/iprule/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteIPRuleCommand represents an intent to permanently remove an IP rule by its unique identifier.
// Once deleted, any traffic previously matched by this rule will fall through to the default policy.
type DeleteIPRuleCommand struct {
	ID uuid.UUID
}

// DeleteIPRuleHandler orchestrates IP rule deletion through the repository layer.
// It enforces a hard-delete strategy — no soft-delete or audit trail is maintained at this level.
// Callers are responsible for authorization checks before invoking this handler.
type DeleteIPRuleHandler struct {
	repo   domain.IPRuleRepository
	logger logger.Log
}

// NewDeleteIPRuleHandler wires up the handler with its required dependencies.
func NewDeleteIPRuleHandler(
	repo domain.IPRuleRepository,
	logger logger.Log,
) *DeleteIPRuleHandler {
	return &DeleteIPRuleHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle performs the deletion of the IP rule identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found, connection failure) to the caller.
func (h *DeleteIPRuleHandler) Handle(ctx context.Context, cmd DeleteIPRuleCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete ip rule: %v", err)
		return err
	}
	return nil
}
