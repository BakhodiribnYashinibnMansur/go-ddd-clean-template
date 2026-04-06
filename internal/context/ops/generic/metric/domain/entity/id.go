package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// MetricID is the typed identifier for a FunctionMetric aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a MetricID is expected).
type MetricID uuid.UUID

// NewMetricID generates a new MetricID backed by a v4 UUID.
func NewMetricID() MetricID { return MetricID(uuid.New()) }

// ParseMetricID parses the canonical UUID string representation of a MetricID.
// It returns an error if s is not a valid UUID.
func ParseMetricID(s string) (MetricID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return MetricID{}, fmt.Errorf("parse metric id %q: %w", s, err)
	}
	return MetricID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id MetricID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id MetricID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the MetricID is the zero value.
func (id MetricID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// MarshalJSON serializes the MetricID as a canonical UUID string.
func (id MetricID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a MetricID from a JSON UUID string.
func (id *MetricID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseMetricID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id MetricID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *MetricID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = MetricID(u)
	return nil
}
