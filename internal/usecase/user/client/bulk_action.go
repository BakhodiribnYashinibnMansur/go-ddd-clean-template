package client

import (
	"context"
	"fmt"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

// BulkAction performs a bulk deactivation or deletion on the given user IDs.
func (uc *UseCase) BulkAction(ctx context.Context, req domain.BulkActionRequest) error {
	uc.logger.Infoc(ctx, "user bulk action started", "action", req.Action, "count", len(req.IDs))

	if len(req.IDs) == 0 {
		return apperrors.New(apperrors.ErrInternal, "ids list is empty")
	}

	var err error
	switch req.Action {
	case "deactivate":
		err = uc.repo.Postgres.User.Client.BulkDeactivate(ctx, req.IDs)
	case "delete":
		err = uc.repo.Postgres.User.Client.BulkDelete(ctx, req.IDs)
	default:
		return apperrors.New(apperrors.ErrInternal, fmt.Sprintf("unknown action: %s", req.Action))
	}

	if err != nil {
		uc.logger.Errorc(ctx, "user bulk action failed", "error", err)
		return apperrors.MapRepoToServiceError(err)
	}

	uc.logger.Infoc(ctx, "user bulk action success", "action", req.Action)
	return nil
}
