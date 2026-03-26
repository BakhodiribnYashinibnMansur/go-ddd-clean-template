package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationSent is a domain event raised when a new notification is created for a user.
// Subscribers can use this to push real-time updates via WebSocket or trigger email/SMS delivery.
type NotificationSent struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	UserID      uuid.UUID
	Title       string
}

func NewNotificationSent(id, userID uuid.UUID, title string) NotificationSent {
	return NotificationSent{
		aggregateID: id,
		occurredAt:  time.Now(),
		UserID:      userID,
		Title:       title,
	}
}

func (e NotificationSent) EventName() string      { return "notification.sent" }
func (e NotificationSent) OccurredAt() time.Time   { return e.occurredAt }
func (e NotificationSent) AggregateID() uuid.UUID  { return e.aggregateID }
