package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// IntegrationID is the typed identifier for an Integration aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where an IntegrationID is expected).
type IntegrationID uuid.UUID

// NewIntegrationID generates a new IntegrationID backed by a v4 UUID.
func NewIntegrationID() IntegrationID { return IntegrationID(uuid.New()) }

// ParseIntegrationID parses the canonical UUID string representation of an IntegrationID.
// It returns an error if s is not a valid UUID.
func ParseIntegrationID(s string) (IntegrationID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return IntegrationID{}, fmt.Errorf("parse integration id %q: %w", s, err)
	}
	return IntegrationID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id IntegrationID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id IntegrationID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the IntegrationID is the zero value.
func (id IntegrationID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// MarshalJSON serializes the IntegrationID as a canonical UUID string.
func (id IntegrationID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a IntegrationID from a JSON UUID string.
func (id *IntegrationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseIntegrationID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id IntegrationID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *IntegrationID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = IntegrationID(u)
	return nil
}
