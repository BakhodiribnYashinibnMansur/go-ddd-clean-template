package command

import (
	"context"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userevent "gct/internal/context/iam/generic/user/domain/event"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ApproveUserCommand holds the input for approving a user.
type ApproveUserCommand struct {
	ID      userentity.UserID
	ActorID uuid.UUID
}

// ApproveUserHandler handles the ApproveUserCommand.
type ApproveUserHandler struct {
	repo     userrepo.UserRepository
	db       shareddomain.DB
	eventBus application.EventBus
	logger   commandLogger
}

// NewApproveUserHandler creates a new ApproveUserHandler.
func NewApproveUserHandler(
	repo userrepo.UserRepository,
	db shareddomain.DB,
	eventBus application.EventBus,
	logger commandLogger,
) *ApproveUserHandler {
	return &ApproveUserHandler{
		repo:     repo,
		db:       db,
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

	user.AddEvent(userevent.NewUserApprovedWithChanges(user.ID(), cmd.ActorID, []userevent.FieldChange{
		{FieldName: "is_approved", OldValue: "false", NewValue: "true"},
	}))

	if err := pgxutil.WithTx(ctx, h.db, func(q shareddomain.Querier) error {
		return h.repo.Update(ctx, q, user)
	}); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "ApproveUser", Entity: "user", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "ApproveUser", Entity: "user", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
