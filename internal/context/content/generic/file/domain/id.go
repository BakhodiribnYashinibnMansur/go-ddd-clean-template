package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// FileID is the typed identifier for a File aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a FileID is expected).
type FileID uuid.UUID

// NewFileID generates a new FileID backed by a v4 UUID.
func NewFileID() FileID { return FileID(uuid.New()) }

// ParseFileID parses the canonical UUID string representation of a FileID.
// It returns an error if s is not a valid UUID.
func ParseFileID(s string) (FileID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return FileID{}, fmt.Errorf("parse file id %q: %w", s, err)
	}
	return FileID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id FileID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id FileID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the FileID is the zero value.
func (id FileID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }
