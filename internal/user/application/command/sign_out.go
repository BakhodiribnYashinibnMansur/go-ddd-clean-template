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

// SignOutCommand holds the input for user sign-out.
type SignOutCommand struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

// SignOutHandler handles the SignOutCommand.
type SignOutHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewSignOutHandler creates a new SignOutHandler.
func NewSignOutHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
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

	user, err := h.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	if err := user.RevokeSession(cmd.SessionID); err != nil {
		return apperrors.MapToServiceError(err)
	}

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "SignOut", Entity: "user", EntityID: cmd.UserID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
