package command

import (
	"context"

	"gct/internal/context/content/notification/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteCommand holds the input for deleting a notification.
type DeleteCommand struct {
	ID domain.NotificationID
}

// DeleteHandler handles the DeleteCommand.
type DeleteHandler struct {
	repo     domain.NotificationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteHandler creates a new DeleteHandler.
func NewDeleteHandler(
	repo domain.NotificationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *DeleteHandler {
	return &DeleteHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the DeleteCommand.
func (h *DeleteHandler) Handle(ctx context.Context, cmd DeleteCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "DeleteNotification", "notification")()

	if err := h.repo.Delete(ctx, cmd.ID.UUID()); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteNotification", Entity: "notification", EntityID: cmd.ID.UUID(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
