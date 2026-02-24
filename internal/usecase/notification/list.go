package notification

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) List(ctx context.Context, filter domain.NotificationFilter) ([]domain.Notification, int64, error) {
	return uc.repo.List(ctx, filter)
}
