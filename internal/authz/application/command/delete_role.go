package command

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// DeleteRoleCommand represents an intent to permanently remove a role from the authorization system.
// Callers must ensure no users are still assigned this role before deletion.
type DeleteRoleCommand struct {
	ID uuid.UUID
}

// DeleteRoleHandler performs hard deletion of a role and emits a RoleDeleted event.
// The event enables downstream consumers (e.g., cache invalidation, user-role cleanup) to react accordingly.
type DeleteRoleHandler struct {
	repo     domain.RoleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteRoleHandler wires dependencies for role deletion.
func NewDeleteRoleHandler(
	repo domain.RoleRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *DeleteRoleHandler {
	return &DeleteRoleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle deletes the role and publishes a RoleDeleted domain event.
// Returns nil on success; propagates repository errors (e.g., not found, FK constraint) to the caller.
func (h *DeleteRoleHandler) Handle(ctx context.Context, cmd DeleteRoleCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteRoleHandler.Handle")
	defer func() { end(err) }()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete role: %v", err)
		return err
	}

	event := domain.NewRoleDeleted(cmd.ID)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
