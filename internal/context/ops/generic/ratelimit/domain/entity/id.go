package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// RateLimitID is the typed identifier for a RateLimit aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a RateLimitID is expected).
type RateLimitID uuid.UUID

// NewRateLimitID generates a new RateLimitID backed by a v4 UUID.
func NewRateLimitID() RateLimitID { return RateLimitID(uuid.New()) }

// ParseRateLimitID parses the canonical UUID string representation of a RateLimitID.
// It returns an error if s is not a valid UUID.
func ParseRateLimitID(s string) (RateLimitID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return RateLimitID{}, fmt.Errorf("parse rate limit id %q: %w", s, err)
	}
	return RateLimitID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id RateLimitID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id RateLimitID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the RateLimitID is the zero value.
func (id RateLimitID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// MarshalJSON serializes the RateLimitID as a canonical UUID string.
func (id RateLimitID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a RateLimitID from a JSON UUID string.
func (id *RateLimitID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseRateLimitID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id RateLimitID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *RateLimitID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = RateLimitID(u)
	return nil
}
