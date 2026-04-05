package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// AnnouncementID is the typed identifier for an Announcement aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where an AnnouncementID is expected).
type AnnouncementID uuid.UUID

// NewAnnouncementID generates a new AnnouncementID backed by a v4 UUID.
func NewAnnouncementID() AnnouncementID { return AnnouncementID(uuid.New()) }

// ParseAnnouncementID parses the canonical UUID string representation of an AnnouncementID.
// It returns an error if s is not a valid UUID.
func ParseAnnouncementID(s string) (AnnouncementID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return AnnouncementID{}, fmt.Errorf("parse announcement id %q: %w", s, err)
	}
	return AnnouncementID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id AnnouncementID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id AnnouncementID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the AnnouncementID is the zero value.
func (id AnnouncementID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// MarshalJSON serializes the AnnouncementID as a canonical UUID string.
func (id AnnouncementID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a AnnouncementID from a JSON UUID string.
func (id *AnnouncementID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseAnnouncementID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id AnnouncementID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *AnnouncementID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = AnnouncementID(u)
	return nil
}
