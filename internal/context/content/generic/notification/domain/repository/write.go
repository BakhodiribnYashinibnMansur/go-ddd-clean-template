package repository

import (
	"context"

	"gct/internal/context/content/generic/notification/domain/entity"
	shareddomain "gct/internal/kernel/domain"
)

// NotificationRepository is the write-side repository for the Notification aggregate.
// Update is used primarily for the MarkAsRead state transition.
type NotificationRepository interface {
	Save(ctx context.Context, q shareddomain.Querier, entity *entity.Notification) error
	FindByID(ctx context.Context, id entity.NotificationID) (*entity.Notification, error)
	Update(ctx context.Context, q shareddomain.Querier, entity *entity.Notification) error
	Delete(ctx context.Context, q shareddomain.Querier, id entity.NotificationID) error
}
