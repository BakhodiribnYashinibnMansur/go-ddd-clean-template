package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// SystemErrorID is the typed identifier for a SystemError aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a SystemErrorID is expected).
type SystemErrorID uuid.UUID

// NewSystemErrorID generates a new SystemErrorID backed by a v4 UUID.
func NewSystemErrorID() SystemErrorID { return SystemErrorID(uuid.New()) }

// ParseSystemErrorID parses the canonical UUID string representation of a SystemErrorID.
// It returns an error if s is not a valid UUID.
func ParseSystemErrorID(s string) (SystemErrorID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SystemErrorID{}, fmt.Errorf("parse system error id %q: %w", s, err)
	}
	return SystemErrorID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id SystemErrorID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id SystemErrorID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the SystemErrorID is the zero value.
func (id SystemErrorID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }
