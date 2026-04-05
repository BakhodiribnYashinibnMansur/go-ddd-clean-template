package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// DataExportID is the typed identifier for a DataExport aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a DataExportID is expected).
type DataExportID uuid.UUID

// NewDataExportID generates a new DataExportID backed by a v4 UUID.
func NewDataExportID() DataExportID { return DataExportID(uuid.New()) }

// ParseDataExportID parses the canonical UUID string representation of a DataExportID.
// It returns an error if s is not a valid UUID.
func ParseDataExportID(s string) (DataExportID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return DataExportID{}, fmt.Errorf("parse data export id %q: %w", s, err)
	}
	return DataExportID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id DataExportID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id DataExportID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the DataExportID is the zero value.
func (id DataExportID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }
