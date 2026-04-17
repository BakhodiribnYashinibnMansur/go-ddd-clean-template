package command

import (
	"context"
	"fmt"

	notifentity "gct/internal/context/content/generic/notification/domain/entity"
	notifrepo "gct/internal/context/content/generic/notification/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"

	"github.com/google/uuid"
)

// CreateCommand holds the input for creating a new notification.
type CreateCommand struct {
	UserID  uuid.UUID
	Title   string
	Message string
	Type    string
}

// CreateHandler handles the CreateCommand.
type CreateHandler struct {
	repo      notifrepo.NotificationRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateHandler creates a new CreateHandler.
func NewCreateHandler(
	repo notifrepo.NotificationRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateHandler {
	return &CreateHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle executes the CreateCommand.
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateNotification", "notification")()

	n, err := notifentity.NewNotification(cmd.UserID, cmd.Title, cmd.Message, cmd.Type)
	if err != nil {
		return fmt.Errorf("create_notification: %w", err)
	}

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Save(ctx, q, n); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateNotification", Entity: "notification", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, n.Events)
}
