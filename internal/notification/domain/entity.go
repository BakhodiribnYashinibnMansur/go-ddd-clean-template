package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// Notification is the aggregate root for per-user notification management.
// Each notification is scoped to a single user via userID. The readAt field tracks read/unread state —
// a nil value means unread. The nType field categorizes the notification (e.g., "INFO", "WARNING", "ALERT").
type Notification struct {
	shared.AggregateRoot
	userID  uuid.UUID
	title   string
	message string
	nType   string
	readAt  *time.Time
}

// NewNotification creates a new Notification aggregate and raises a NotificationSent event.
func NewNotification(userID uuid.UUID, title, message, nType string) *Notification {
	n := &Notification{
		AggregateRoot: shared.NewAggregateRoot(),
		userID:        userID,
		title:         title,
		message:       message,
		nType:         nType,
	}
	n.AddEvent(NewNotificationSent(n.ID(), userID, title))
	return n
}

// ReconstructNotification rebuilds a Notification aggregate from persisted data. No events are raised.
func ReconstructNotification(
	id uuid.UUID,
	createdAt time.Time,
	userID uuid.UUID,
	title, message, nType string,
	readAt *time.Time,
) *Notification {
	return &Notification{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, createdAt, nil),
		userID:        userID,
		title:         title,
		message:       message,
		nType:         nType,
		readAt:        readAt,
	}
}

// MarkAsRead sets the readAt timestamp to now, transitioning the notification to "read" state.
// This operation is idempotent — calling it on an already-read notification overwrites the timestamp.
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.readAt = &now
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (n *Notification) UserID() uuid.UUID   { return n.userID }
func (n *Notification) Title() string       { return n.title }
func (n *Notification) Message() string     { return n.message }
func (n *Notification) Type() string        { return n.nType }
func (n *Notification) ReadAt() *time.Time  { return n.readAt }
