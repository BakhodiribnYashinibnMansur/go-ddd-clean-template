package notification

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	List(ctx context.Context, filter domain.NotificationFilter) ([]domain.Notification, int64, error)
	Update(ctx context.Context, n *domain.Notification) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UseCaseI interface {
	Create(ctx context.Context, req domain.CreateNotificationRequest) (*domain.Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	List(ctx context.Context, filter domain.NotificationFilter) ([]domain.Notification, int64, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateNotificationRequest) (*domain.Notification, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
