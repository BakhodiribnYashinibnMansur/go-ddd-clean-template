package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// IPRuleID is the typed identifier for an IPRule aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where an IPRuleID is expected).
type IPRuleID uuid.UUID

// NewIPRuleID generates a new IPRuleID backed by a v4 UUID.
func NewIPRuleID() IPRuleID { return IPRuleID(uuid.New()) }

// ParseIPRuleID parses the canonical UUID string representation of an IPRuleID.
// It returns an error if s is not a valid UUID.
func ParseIPRuleID(s string) (IPRuleID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return IPRuleID{}, fmt.Errorf("parse ip rule id %q: %w", s, err)
	}
	return IPRuleID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id IPRuleID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id IPRuleID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the IPRuleID is the zero value.
func (id IPRuleID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// MarshalJSON serializes the IPRuleID as a canonical UUID string.
func (id IPRuleID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses an IPRuleID from a JSON UUID string.
func (id *IPRuleID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseIPRuleID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id IPRuleID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *IPRuleID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = IPRuleID(u)
	return nil
}
