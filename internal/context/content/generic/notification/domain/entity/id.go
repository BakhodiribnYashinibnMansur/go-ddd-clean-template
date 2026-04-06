package entity

import (
	"database/sql/driver"
	"encoding/json"
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

// MarshalJSON serializes the NotificationID as a canonical UUID string.
func (id NotificationID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a NotificationID from a JSON UUID string.
func (id *NotificationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseNotificationID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id NotificationID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *NotificationID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = NotificationID(u)
	return nil
}
