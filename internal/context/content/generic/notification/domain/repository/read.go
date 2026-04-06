package repository

import (
	"context"
	"time"

	"gct/internal/context/content/generic/notification/domain/entity"

	"github.com/google/uuid"
)

// NotificationFilter carries optional filtering parameters for listing notifications.
// The Unread flag, when set to true, restricts results to notifications where readAt is nil.
type NotificationFilter struct {
	UserID *uuid.UUID
	Type   *string
	Unread *bool
	Limit  int64
	Offset int64
}

// NotificationView is a read-model projection optimized for query responses.
// ReadAt being nil indicates an unread notification in the UI layer.
type NotificationView struct {
	ID        entity.NotificationID `json:"id"`
	UserID    uuid.UUID             `json:"user_id"`
	Title     string                `json:"title"`
	Message   string                `json:"message"`
	Type      string                `json:"type"`
	ReadAt    *time.Time            `json:"read_at,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
}

// NotificationReadRepository is the read-side repository returning projected views.
// Implementations must return ErrNotificationNotFound when FindByID yields no result.
type NotificationReadRepository interface {
	FindByID(ctx context.Context, id entity.NotificationID) (*NotificationView, error)
	List(ctx context.Context, filter NotificationFilter) ([]*NotificationView, int64, error)
}
