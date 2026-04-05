package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// TranslationID is the typed identifier for a Translation aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a TranslationID is expected).
type TranslationID uuid.UUID

// NewTranslationID generates a new TranslationID backed by a v4 UUID.
func NewTranslationID() TranslationID { return TranslationID(uuid.New()) }

// ParseTranslationID parses the canonical UUID string representation of a TranslationID.
// It returns an error if s is not a valid UUID.
func ParseTranslationID(s string) (TranslationID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TranslationID{}, fmt.Errorf("parse translation id %q: %w", s, err)
	}
	return TranslationID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id TranslationID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id TranslationID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the TranslationID is the zero value.
func (id TranslationID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// MarshalJSON serializes the TranslationID as a canonical UUID string.
func (id TranslationID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a TranslationID from a JSON UUID string.
func (id *TranslationID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseTranslationID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id TranslationID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *TranslationID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = TranslationID(u)
	return nil
}
