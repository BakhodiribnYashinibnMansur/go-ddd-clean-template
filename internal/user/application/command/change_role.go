package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// ChangeRoleCommand holds the input for changing a user's role.
type ChangeRoleCommand struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

// ChangeRoleHandler handles the ChangeRoleCommand.
type ChangeRoleHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewChangeRoleHandler creates a new ChangeRoleHandler.
func NewChangeRoleHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *ChangeRoleHandler {
	return &ChangeRoleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the ChangeRoleCommand.
func (h *ChangeRoleHandler) Handle(ctx context.Context, cmd ChangeRoleCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ChangeRoleHandler.Handle")
	defer func() { end(err) }()

	user, err := h.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	user.ChangeRole(cmd.RoleID)

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorf("failed to save role change: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Errorf("failed to publish role change events: %v", err)
	}

	return nil
}
