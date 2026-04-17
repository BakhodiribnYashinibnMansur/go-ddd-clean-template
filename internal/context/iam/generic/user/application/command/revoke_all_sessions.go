package command

import (
	"context"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// RevokeAllSessionsCommand holds the input for revoking all user sessions.
type RevokeAllSessionsCommand struct {
	UserID userentity.UserID
}

// RevokeAllSessionsHandler handles the RevokeAllSessionsCommand.
type RevokeAllSessionsHandler struct {
	repo     userrepo.UserRepository
	db       shareddomain.DB
	eventBus application.EventBus
	logger   commandLogger
}

// NewRevokeAllSessionsHandler creates a new RevokeAllSessionsHandler.
func NewRevokeAllSessionsHandler(
	repo userrepo.UserRepository,
	db shareddomain.DB,
	eventBus application.EventBus,
	logger commandLogger,
) *RevokeAllSessionsHandler {
	return &RevokeAllSessionsHandler{
		repo:     repo,
		db:       db,
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

	if err := pgxutil.WithTx(ctx, h.db, func(q shareddomain.Querier) error {
		return h.repo.Update(ctx, q, user)
	}); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "RevokeAllSessions", Entity: "user", EntityID: cmd.UserID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
