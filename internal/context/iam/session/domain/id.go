// Package domain defines the types owned by the Session read-only bounded context.
//
// The Session BC is a read-side projection BC: the session aggregate itself is
// owned by the User BC. However, the Session BC exposes its own typed
// identifiers for its query surface so that callers inside this BC and its
// compile-time boundary get the same safety guarantees as the user BC.
//
// Cross-BC references at the wire/contract level remain raw uuid.UUID.
package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// SessionID is the typed identifier for a Session read-model within the
// Session BC. It mirrors the identity managed by the User BC.
type SessionID uuid.UUID

// NewSessionID generates a new SessionID backed by a v4 UUID.
func NewSessionID() SessionID { return SessionID(uuid.New()) }

// ParseSessionID parses the canonical UUID string representation of a SessionID.
func ParseSessionID(s string) (SessionID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SessionID{}, fmt.Errorf("parse session id %q: %w", s, err)
	}
	return SessionID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id SessionID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id SessionID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the SessionID is the zero value.
func (id SessionID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// UserID is the typed identifier used by the Session BC to refer to the User
// aggregate that owns a session. It is a local type for compile-time safety;
// it is not a cross-BC reference to the User BC's own UserID type.
type UserID uuid.UUID

// NewUserID generates a new UserID backed by a v4 UUID.
func NewUserID() UserID { return UserID(uuid.New()) }

// ParseUserID parses the canonical UUID string representation of a UserID.
func ParseUserID(s string) (UserID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, fmt.Errorf("parse user id %q: %w", s, err)
	}
	return UserID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id UserID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id UserID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the UserID is the zero value.
func (id UserID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }
