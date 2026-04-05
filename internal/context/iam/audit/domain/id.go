package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// AuditLogID is the typed identifier for an AuditLog aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing an EndpointHistoryID where an AuditLogID is expected).
type AuditLogID uuid.UUID

// NewAuditLogID generates a new AuditLogID backed by a v4 UUID.
func NewAuditLogID() AuditLogID { return AuditLogID(uuid.New()) }

// ParseAuditLogID parses the canonical UUID string representation of an AuditLogID.
// It returns an error if s is not a valid UUID.
func ParseAuditLogID(s string) (AuditLogID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return AuditLogID{}, fmt.Errorf("parse audit log id %q: %w", s, err)
	}
	return AuditLogID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id AuditLogID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id AuditLogID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the AuditLogID is the zero value.
func (id AuditLogID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// EndpointHistoryID is the typed identifier for an EndpointHistory entity.
type EndpointHistoryID uuid.UUID

// NewEndpointHistoryID generates a new EndpointHistoryID backed by a v4 UUID.
func NewEndpointHistoryID() EndpointHistoryID { return EndpointHistoryID(uuid.New()) }

// ParseEndpointHistoryID parses the canonical UUID string representation of an EndpointHistoryID.
func ParseEndpointHistoryID(s string) (EndpointHistoryID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return EndpointHistoryID{}, fmt.Errorf("parse endpoint history id %q: %w", s, err)
	}
	return EndpointHistoryID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id EndpointHistoryID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id EndpointHistoryID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the EndpointHistoryID is the zero value.
func (id EndpointHistoryID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }
