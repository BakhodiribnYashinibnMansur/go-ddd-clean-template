package command

import (
	"context"

	"gct/internal/context/iam/user/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// SignOutCommand holds the input for user sign-out.
type SignOutCommand struct {
	UserID    domain.UserID
	SessionID domain.SessionID
}

// SignOutHandler handles the SignOutCommand.
type SignOutHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   commandLogger
}

// NewSignOutHandler creates a new SignOutHandler.
func NewSignOutHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger commandLogger,
) *SignOutHandler {
	return &SignOutHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the SignOutCommand.
func (h *SignOutHandler) Handle(ctx context.Context, cmd SignOutCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "SignOutHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "SignOut", "user")()

	user, err := h.repo.FindByID(ctx, cmd.UserID.UUID())
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	if err := user.RevokeSession(cmd.SessionID.UUID()); err != nil {
		return apperrors.MapToServiceError(err)
	}

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "SignOut", Entity: "user", EntityID: cmd.UserID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
