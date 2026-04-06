package repository

import (
	"context"

	"gct/internal/context/content/generic/notification/domain/entity"
)

// NotificationRepository is the write-side repository for the Notification aggregate.
// Update is used primarily for the MarkAsRead state transition.
type NotificationRepository interface {
	Save(ctx context.Context, entity *entity.Notification) error
	FindByID(ctx context.Context, id entity.NotificationID) (*entity.Notification, error)
	Update(ctx context.Context, entity *entity.Notification) error
	Delete(ctx context.Context, id entity.NotificationID) error
}
