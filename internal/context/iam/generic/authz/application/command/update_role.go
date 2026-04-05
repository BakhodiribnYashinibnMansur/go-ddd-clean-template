package command

import (
	"context"

	"gct/internal/context/iam/generic/authz/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// UpdateRoleCommand represents a partial update to an existing role.
// Nil pointer fields are left unchanged, enabling callers to modify name or description independently.
type UpdateRoleCommand struct {
	ID          domain.RoleID
	Name        *string
	Description *string
}

// UpdateRoleHandler applies partial modifications to an existing role and emits domain events.
// The handler follows a fetch-mutate-persist pattern; event publish failures are non-fatal.
type UpdateRoleHandler struct {
	repo     domain.RoleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateRoleHandler wires dependencies for role updates.
func NewUpdateRoleHandler(
	repo domain.RoleRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateRoleHandler {
	return &UpdateRoleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle fetches the role by ID, applies non-nil field updates, persists, and publishes events.
// Returns a repository error if the role is not found or the update fails.
func (h *UpdateRoleHandler) Handle(ctx context.Context, cmd UpdateRoleCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateRoleHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateRole", "role")()

	role, err := h.repo.FindByID(ctx, cmd.ID.UUID())
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	if cmd.Name != nil {
		role.Rename(*cmd.Name)
	}
	if cmd.Description != nil {
		role.SetDescription(cmd.Description)
	}

	if err := h.repo.Update(ctx, role); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "UpdateRole", Entity: "role", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, role.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateRole", Entity: "role", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
	}

	return nil
}
