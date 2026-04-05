package command

import (
	"context"
	"fmt"

	"gct/internal/context/iam/generic/user/domain"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

const (
	BulkActionActivate   = "activate"
	BulkActionDeactivate = "deactivate"
	BulkActionDelete     = "delete"
)

// BulkActionCommand holds the input for performing a bulk action on users.
type BulkActionCommand struct {
	IDs    []domain.UserID
	Action string
}

// BulkActionHandler handles the BulkActionCommand.
type BulkActionHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   commandLogger
}

// NewBulkActionHandler creates a new BulkActionHandler.
func NewBulkActionHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger commandLogger,
) *BulkActionHandler {
	return &BulkActionHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the BulkActionCommand.
func (h *BulkActionHandler) Handle(ctx context.Context, cmd BulkActionCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "BulkActionHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "BulkAction", "user")()

	for _, id := range cmd.IDs {
		user, err := h.repo.FindByID(ctx, id)
		if err != nil {
			h.logger.Warnc(ctx, "bulk action: user find failed", logger.F{Op: "BulkAction", Entity: "user", EntityID: id.String(), Err: err}.KV()...)
			continue
		}

		switch cmd.Action {
		case BulkActionActivate:
			user.Activate()
		case BulkActionDeactivate:
			user.Deactivate()
		case BulkActionDelete:
			user.Deactivate()
			user.SoftDelete()
		default:
			return fmt.Errorf("unknown bulk action: %s", cmd.Action)
		}

		if err := h.repo.Update(ctx, user); err != nil {
			h.logger.Errorc(ctx, "bulk action: repository update failed", logger.F{Op: "BulkAction", Entity: "user", EntityID: id.String(), Err: err}.KV()...)
			continue
		}

		if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
			h.logger.Warnc(ctx, "bulk action: event publish failed", logger.F{Op: "BulkAction", Entity: "user", EntityID: id.String(), Err: err}.KV()...)
		}
	}

	return nil
}
