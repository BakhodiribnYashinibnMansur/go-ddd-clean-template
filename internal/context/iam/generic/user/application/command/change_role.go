package command

import (
	"context"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userevent "gct/internal/context/iam/generic/user/domain/event"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ChangeRoleCommand holds the input for changing a user's role.
type ChangeRoleCommand struct {
	UserID  userentity.UserID
	ActorID uuid.UUID
	// RoleID is owned by the Authz BC and stays as uuid.UUID at this boundary.
	RoleID uuid.UUID
}

// ChangeRoleHandler handles the ChangeRoleCommand.
type ChangeRoleHandler struct {
	repo     userrepo.UserRepository
	eventBus application.EventBus
	logger   commandLogger
}

// NewChangeRoleHandler creates a new ChangeRoleHandler.
func NewChangeRoleHandler(
	repo userrepo.UserRepository,
	eventBus application.EventBus,
	logger commandLogger,
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

	oldRoleID := ""
	if user.RoleID() != nil {
		oldRoleID = user.RoleID().String()
	}

	user.ChangeRole(cmd.RoleID)

	user.AddEvent(userevent.NewRoleChangedWithChanges(user.ID(), cmd.ActorID, []userevent.FieldChange{
		{FieldName: "role_id", OldValue: oldRoleID, NewValue: cmd.RoleID.String()},
	}))

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "ChangeRole", Entity: "user", EntityID: cmd.UserID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "ChangeRole", Entity: "user", EntityID: cmd.UserID, Err: err}.KV()...)
	}

	return nil
}
