package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// ErrorCodeID is the typed identifier for an ErrorCode aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where an ErrorCodeID is expected).
type ErrorCodeID uuid.UUID

// NewErrorCodeID generates a new ErrorCodeID backed by a v4 UUID.
func NewErrorCodeID() ErrorCodeID { return ErrorCodeID(uuid.New()) }

// ParseErrorCodeID parses the canonical UUID string representation of an ErrorCodeID.
// It returns an error if s is not a valid UUID.
func ParseErrorCodeID(s string) (ErrorCodeID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ErrorCodeID{}, fmt.Errorf("parse error code id %q: %w", s, err)
	}
	return ErrorCodeID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id ErrorCodeID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id ErrorCodeID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the ErrorCodeID is the zero value.
func (id ErrorCodeID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }
