package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationSent is raised when a new notification is created.
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
