package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// UserID is the typed identifier for a User aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing a SessionID where a UserID is expected).
type UserID uuid.UUID

// NewUserID generates a new UserID backed by a v4 UUID.
func NewUserID() UserID { return UserID(uuid.New()) }

// ParseUserID parses the canonical UUID string representation of a UserID.
// It returns an error if s is not a valid UUID.
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

// SessionID is the typed identifier for a Session entity owned by the User aggregate.
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
