package domain

import (
	"database/sql/driver"
	"encoding/json"
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

// MarshalJSON serializes the FileID as a canonical UUID string.
func (id FileID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a FileID from a JSON UUID string.
func (id *FileID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseFileID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id FileID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *FileID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = FileID(u)
	return nil
}
