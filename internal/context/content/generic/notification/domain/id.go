package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// NotificationID is the typed identifier for a Notification aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a NotificationID is expected).
type NotificationID uuid.UUID

// NewNotificationID generates a new NotificationID backed by a v4 UUID.
func NewNotificationID() NotificationID { return NotificationID(uuid.New()) }

// ParseNotificationID parses the canonical UUID string representation of a NotificationID.
// It returns an error if s is not a valid UUID.
func ParseNotificationID(s string) (NotificationID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return NotificationID{}, fmt.Errorf("parse notification id %q: %w", s, err)
	}
	return NotificationID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id NotificationID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id NotificationID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the NotificationID is the zero value.
func (id NotificationID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }
