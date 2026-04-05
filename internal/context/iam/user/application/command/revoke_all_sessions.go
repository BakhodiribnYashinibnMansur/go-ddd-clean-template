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

// RevokeAllSessionsCommand holds the input for revoking all user sessions.
type RevokeAllSessionsCommand struct {
	UserID uuid.UUID
}

// RevokeAllSessionsHandler handles the RevokeAllSessionsCommand.
type RevokeAllSessionsHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewRevokeAllSessionsHandler creates a new RevokeAllSessionsHandler.
func NewRevokeAllSessionsHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *RevokeAllSessionsHandler {
	return &RevokeAllSessionsHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the RevokeAllSessionsCommand.
func (h *RevokeAllSessionsHandler) Handle(ctx context.Context, cmd RevokeAllSessionsCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "RevokeAllSessionsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "RevokeAllSessions", "user")()

	user, err := h.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	user.RevokeAllSessions()

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "RevokeAllSessions", Entity: "user", EntityID: cmd.UserID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
