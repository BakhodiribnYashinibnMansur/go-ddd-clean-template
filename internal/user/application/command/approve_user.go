package command

import (
	"context"

	"gct/internal/shared/application"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// ApproveUserCommand holds the input for approving a user.
type ApproveUserCommand struct {
	ID uuid.UUID
}

// ApproveUserHandler handles the ApproveUserCommand.
type ApproveUserHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewApproveUserHandler creates a new ApproveUserHandler.
func NewApproveUserHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *ApproveUserHandler {
	return &ApproveUserHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the ApproveUserCommand.
func (h *ApproveUserHandler) Handle(ctx context.Context, cmd ApproveUserCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ApproveUserHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ApproveUser", "user")()

	user, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	user.Approve()

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "ApproveUser", Entity: "user", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "ApproveUser", Entity: "user", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
