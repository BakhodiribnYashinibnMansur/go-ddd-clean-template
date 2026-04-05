package domain

import (
	"database/sql/driver"
	"encoding/json"
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

// MarshalJSON serializes the UserID as a canonical UUID string.
func (id UserID) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }

// UnmarshalJSON parses a UserID from a JSON UUID string.
func (id *UserID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseUserID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id UserID) Value() (driver.Value, error) { return uuid.UUID(id).Value() }

// Scan implements sql.Scanner for SQL driver interop.
func (id *UserID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = UserID(u)
	return nil
}

// SessionID is the typed identifier for a Session entity owned by the User aggregate.
type SessionID uuid.UUID

// NilSessionID is the zero-valued SessionID, used to signal "no session" from
// idempotent operations (e.g. RevokeOldestActiveSession when nothing matched).
var NilSessionID = SessionID(uuid.Nil)

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

// MarshalJSON serializes the SessionID as a canonical UUID string.
func (id SessionID) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }

// UnmarshalJSON parses a SessionID from a JSON UUID string.
func (id *SessionID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseSessionID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id SessionID) Value() (driver.Value, error) { return uuid.UUID(id).Value() }

// Scan implements sql.Scanner for SQL driver interop.
func (id *SessionID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = SessionID(u)
	return nil
}
