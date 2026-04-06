package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// FeatureFlagID is the typed identifier for a FeatureFlag aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a FeatureFlagID is expected).
type FeatureFlagID uuid.UUID

// NewFeatureFlagID generates a new FeatureFlagID backed by a v4 UUID.
func NewFeatureFlagID() FeatureFlagID { return FeatureFlagID(uuid.New()) }

// ParseFeatureFlagID parses the canonical UUID string representation of a FeatureFlagID.
// It returns an error if s is not a valid UUID.
func ParseFeatureFlagID(s string) (FeatureFlagID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return FeatureFlagID{}, fmt.Errorf("parse feature flag id %q: %w", s, err)
	}
	return FeatureFlagID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id FeatureFlagID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id FeatureFlagID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the FeatureFlagID is the zero value.
func (id FeatureFlagID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// RuleGroupID is the typed identifier for a RuleGroup entity owned by the
// FeatureFlag aggregate.
type RuleGroupID uuid.UUID

// NewRuleGroupID generates a new RuleGroupID backed by a v4 UUID.
func NewRuleGroupID() RuleGroupID { return RuleGroupID(uuid.New()) }

// ParseRuleGroupID parses the canonical UUID string representation of a RuleGroupID.
// It returns an error if s is not a valid UUID.
func ParseRuleGroupID(s string) (RuleGroupID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return RuleGroupID{}, fmt.Errorf("parse rule group id %q: %w", s, err)
	}
	return RuleGroupID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id RuleGroupID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id RuleGroupID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the RuleGroupID is the zero value.
func (id RuleGroupID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// MarshalJSON serializes the FeatureFlagID as a canonical UUID string.
func (id FeatureFlagID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a FeatureFlagID from a JSON UUID string.
func (id *FeatureFlagID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseFeatureFlagID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id FeatureFlagID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *FeatureFlagID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = FeatureFlagID(u)
	return nil
}
