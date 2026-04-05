package domain

import (
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
