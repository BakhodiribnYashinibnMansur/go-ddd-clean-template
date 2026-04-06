package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// SystemErrorID is the typed identifier for a SystemError aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a SystemErrorID is expected).
type SystemErrorID uuid.UUID

// NewSystemErrorID generates a new SystemErrorID backed by a v4 UUID.
func NewSystemErrorID() SystemErrorID { return SystemErrorID(uuid.New()) }

// ParseSystemErrorID parses the canonical UUID string representation of a SystemErrorID.
// It returns an error if s is not a valid UUID.
func ParseSystemErrorID(s string) (SystemErrorID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SystemErrorID{}, fmt.Errorf("parse system error id %q: %w", s, err)
	}
	return SystemErrorID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id SystemErrorID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id SystemErrorID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the SystemErrorID is the zero value.
func (id SystemErrorID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// MarshalJSON serializes the SystemErrorID as a canonical UUID string.
func (id SystemErrorID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a SystemErrorID from a JSON UUID string.
func (id *SystemErrorID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseSystemErrorID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id SystemErrorID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *SystemErrorID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = SystemErrorID(u)
	return nil
}
