package webhook

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, w *domain.Webhook) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error)
	List(ctx context.Context, filter domain.WebhookFilter) ([]domain.Webhook, int64, error)
	Update(ctx context.Context, w *domain.Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UseCaseI interface {
	Create(ctx context.Context, req domain.CreateWebhookRequest) (*domain.Webhook, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error)
	List(ctx context.Context, filter domain.WebhookFilter) ([]domain.Webhook, int64, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateWebhookRequest) (*domain.Webhook, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Test(ctx context.Context, id uuid.UUID) error
}
