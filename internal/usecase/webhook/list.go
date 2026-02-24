package webhook

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) List(ctx context.Context, filter domain.WebhookFilter) ([]domain.Webhook, int64, error) {
	return uc.repo.List(ctx, filter)
}
