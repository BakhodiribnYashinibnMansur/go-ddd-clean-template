package command

import (
	"context"

	"gct/internal/context/iam/user/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// RevokeAllSessionsCommand holds the input for revoking all user sessions.
type RevokeAllSessionsCommand struct {
	UserID domain.UserID
}

// RevokeAllSessionsHandler handles the RevokeAllSessionsCommand.
type RevokeAllSessionsHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   commandLogger
}

// NewRevokeAllSessionsHandler creates a new RevokeAllSessionsHandler.
func NewRevokeAllSessionsHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger commandLogger,
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

	user, err := h.repo.FindByID(ctx, cmd.UserID.UUID())
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	user.RevokeAllSessions()

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "RevokeAllSessions", Entity: "user", EntityID: cmd.UserID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
