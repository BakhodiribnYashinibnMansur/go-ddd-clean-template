package command

import (
	"context"
	"fmt"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

const (
	BulkActionActivate   = "activate"
	BulkActionDeactivate = "deactivate"
	BulkActionDelete     = "delete"
)

// BulkActionCommand holds the input for performing a bulk action on users.
type BulkActionCommand struct {
	IDs    []uuid.UUID
	Action string
}

// BulkActionHandler handles the BulkActionCommand.
type BulkActionHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewBulkActionHandler creates a new BulkActionHandler.
func NewBulkActionHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger logger.Log,
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

	for _, id := range cmd.IDs {
		user, err := h.repo.FindByID(ctx, id)
		if err != nil {
			h.logger.Errorf("bulk action: failed to find user %s: %v", id, err)
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
			h.logger.Errorf("bulk action: failed to update user %s: %v", id, err)
			continue
		}

		if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
			h.logger.Errorf("bulk action: failed to publish events for user %s: %v", id, err)
		}
	}

	return nil
}
