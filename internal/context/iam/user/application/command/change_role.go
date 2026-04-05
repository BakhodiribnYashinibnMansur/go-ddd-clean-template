package command

import (
	"context"

	"gct/internal/platform/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
	"gct/internal/context/iam/user/domain"

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
	defer logger.SlowOp(h.logger, ctx, "ChangeRole", "user")()

	user, err := h.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	user.ChangeRole(cmd.RoleID)

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "ChangeRole", Entity: "user", EntityID: cmd.UserID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "ChangeRole", Entity: "user", EntityID: cmd.UserID, Err: err}.KV()...)
	}

	return nil
}
