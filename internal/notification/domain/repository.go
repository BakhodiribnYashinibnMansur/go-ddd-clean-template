package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// NotificationFilter carries filtering parameters for listing notifications.
type NotificationFilter struct {
	UserID *uuid.UUID
	Type   *string
	Unread *bool
	Limit  int64
	Offset int64
}

// NotificationView is a read-model DTO for notifications.
type NotificationView struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Type      string     `json:"type"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// NotificationRepository is the write-side repository for the Notification aggregate.
type NotificationRepository interface {
	Save(ctx context.Context, entity *Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	Update(ctx context.Context, entity *Notification) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// NotificationReadRepository is the read-side repository returning projected views.
type NotificationReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*NotificationView, error)
	List(ctx context.Context, filter NotificationFilter) ([]*NotificationView, int64, error)
}
