package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// Typed identifiers for authz aggregates.
//
// The authz bounded context owns four distinct aggregates — Role, Permission,
// Policy, and Scope (the last identified by a composite key, not a UUID) —
// plus their cross-references. Using distinct Go types for each UUID-based ID
// prevents call sites from accidentally passing, for example, a PermissionID
// where a RoleID is expected, a class of bug that is otherwise impossible for
// the compiler to catch because all raw uuid.UUID values are interchangeable.

// RoleID is the typed identifier for a Role aggregate.
type RoleID uuid.UUID

// NewRoleID generates a new RoleID backed by a v4 UUID.
func NewRoleID() RoleID { return RoleID(uuid.New()) }

// ParseRoleID parses the canonical UUID string representation of a RoleID.
// It returns an error if s is not a valid UUID.
func ParseRoleID(s string) (RoleID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return RoleID{}, fmt.Errorf("parse role id %q: %w", s, err)
	}
	return RoleID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id RoleID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id RoleID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the RoleID is the zero value.
func (id RoleID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// PermissionID is the typed identifier for a Permission aggregate.
type PermissionID uuid.UUID

// NewPermissionID generates a new PermissionID backed by a v4 UUID.
func NewPermissionID() PermissionID { return PermissionID(uuid.New()) }

// ParsePermissionID parses the canonical UUID string representation of a PermissionID.
// It returns an error if s is not a valid UUID.
func ParsePermissionID(s string) (PermissionID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return PermissionID{}, fmt.Errorf("parse permission id %q: %w", s, err)
	}
	return PermissionID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id PermissionID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id PermissionID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the PermissionID is the zero value.
func (id PermissionID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// PolicyID is the typed identifier for a Policy aggregate.
type PolicyID uuid.UUID

// NewPolicyID generates a new PolicyID backed by a v4 UUID.
func NewPolicyID() PolicyID { return PolicyID(uuid.New()) }

// ParsePolicyID parses the canonical UUID string representation of a PolicyID.
// It returns an error if s is not a valid UUID.
func ParsePolicyID(s string) (PolicyID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return PolicyID{}, fmt.Errorf("parse policy id %q: %w", s, err)
	}
	return PolicyID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id PolicyID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id PolicyID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the PolicyID is the zero value.
func (id PolicyID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// ScopeID is the typed identifier for a Scope aggregate.
//
// Scope is identified in persistence by the composite (path, method) key, but
// a UUID-backed surrogate ID is defined here for completeness and for any
// future migration where scopes become first-class identifiable entities.
type ScopeID uuid.UUID

// NewScopeID generates a new ScopeID backed by a v4 UUID.
func NewScopeID() ScopeID { return ScopeID(uuid.New()) }

// ParseScopeID parses the canonical UUID string representation of a ScopeID.
// It returns an error if s is not a valid UUID.
func ParseScopeID(s string) (ScopeID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ScopeID{}, fmt.Errorf("parse scope id %q: %w", s, err)
	}
	return ScopeID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id ScopeID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id ScopeID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the ScopeID is the zero value.
func (id ScopeID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }
