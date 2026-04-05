package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// UserSettingID is the typed identifier for a UserSetting aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a UserSettingID is expected).
type UserSettingID uuid.UUID

// NewUserSettingID generates a new UserSettingID backed by a v4 UUID.
func NewUserSettingID() UserSettingID { return UserSettingID(uuid.New()) }

// ParseUserSettingID parses the canonical UUID string representation of a UserSettingID.
// It returns an error if s is not a valid UUID.
func ParseUserSettingID(s string) (UserSettingID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UserSettingID{}, fmt.Errorf("parse user setting id %q: %w", s, err)
	}
	return UserSettingID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id UserSettingID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id UserSettingID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the UserSettingID is the zero value.
func (id UserSettingID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }
