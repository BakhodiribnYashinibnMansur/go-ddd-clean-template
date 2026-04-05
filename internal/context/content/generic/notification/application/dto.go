package application

import (
	"time"

	"gct/internal/context/content/generic/notification/domain"

	"github.com/google/uuid"
)

// NotificationView is a read-model DTO returned by query handlers.
type NotificationView struct {
	ID        domain.NotificationID `json:"id"`
	UserID    uuid.UUID             `json:"user_id"`
	Title     string                `json:"title"`
	Message   string                `json:"message"`
	Type      string                `json:"type"`
	ReadAt    *time.Time            `json:"read_at,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
}
