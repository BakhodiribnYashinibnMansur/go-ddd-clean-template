package command

import (
	"context"
	"fmt"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
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
	IDs    []userentity.UserID
	Action string
}

// BulkActionResult summarises the outcome of a bulk operation so the caller
// can distinguish full success from partial failure.
type BulkActionResult struct {
	Succeeded int
	Failed    int
	Errors    []string
}

// BulkActionHandler handles the BulkActionCommand.
type BulkActionHandler struct {
	repo     userrepo.UserRepository
	db       shareddomain.DB
	eventBus application.EventBus
	logger   commandLogger
}

// NewBulkActionHandler creates a new BulkActionHandler.
func NewBulkActionHandler(
	repo userrepo.UserRepository,
	db shareddomain.DB,
	eventBus application.EventBus,
	logger commandLogger,
) *BulkActionHandler {
	return &BulkActionHandler{
		repo:     repo,
		db:       db,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the BulkActionCommand and returns a result summarising
// successes and failures so that callers never silently lose errors.
func (h *BulkActionHandler) Handle(ctx context.Context, cmd BulkActionCommand) (_ *BulkActionResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "BulkActionHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "BulkAction", "user")()

	result := &BulkActionResult{}

	for _, id := range cmd.IDs {
		user, err := h.repo.FindByID(ctx, id)
		if err != nil {
			h.logger.Warnc(ctx, "bulk action: user find failed", logger.F{Op: "BulkAction", Entity: "user", EntityID: id.String(), Err: err}.KV()...)
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: find failed", id))
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
			return nil, fmt.Errorf("unknown bulk action: %s", cmd.Action)
		}

		if err := pgxutil.WithTx(ctx, h.db, func(q shareddomain.Querier) error {
			return h.repo.Update(ctx, q, user)
		}); err != nil {
			h.logger.Errorc(ctx, "bulk action: repository update failed", logger.F{Op: "BulkAction", Entity: "user", EntityID: id.String(), Err: err}.KV()...)
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: update failed", id))
			continue
		}

		if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
			h.logger.Warnc(ctx, "bulk action: event publish failed", logger.F{Op: "BulkAction", Entity: "user", EntityID: id.String(), Err: err}.KV()...)
		}

		result.Succeeded++
	}

	if result.Failed > 0 {
		return result, fmt.Errorf("bulk action: %d/%d operations failed", result.Failed, result.Failed+result.Succeeded)
	}

	return result, nil
}
