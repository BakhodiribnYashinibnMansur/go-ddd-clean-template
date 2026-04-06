package command

import (
	"context"

	notifentity "gct/internal/context/content/generic/notification/domain/entity"
	notifrepo "gct/internal/context/content/generic/notification/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteCommand holds the input for deleting a notification.
type DeleteCommand struct {
	ID notifentity.NotificationID
}

// DeleteHandler handles the DeleteCommand.
type DeleteHandler struct {
	repo     notifrepo.NotificationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteHandler creates a new DeleteHandler.
func NewDeleteHandler(
	repo notifrepo.NotificationRepository,
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

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteNotification", Entity: "notification", EntityID: cmd.ID.String(), Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	return nil
}
