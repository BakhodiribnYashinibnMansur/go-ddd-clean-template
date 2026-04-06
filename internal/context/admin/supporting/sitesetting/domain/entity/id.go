package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// SiteSettingID is the typed identifier for a SiteSetting aggregate.
// Typed IDs prevent the compiler from mixing up entity identifiers at call
// sites (e.g. passing some other ID where a SiteSettingID is expected).
type SiteSettingID uuid.UUID

// NewSiteSettingID generates a new SiteSettingID backed by a v4 UUID.
func NewSiteSettingID() SiteSettingID { return SiteSettingID(uuid.New()) }

// ParseSiteSettingID parses the canonical UUID string representation of a SiteSettingID.
// It returns an error if s is not a valid UUID.
func ParseSiteSettingID(s string) (SiteSettingID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SiteSettingID{}, fmt.Errorf("parse site setting id %q: %w", s, err)
	}
	return SiteSettingID(id), nil
}

// UUID returns the underlying uuid.UUID for interop with repository / UUID-based APIs.
func (id SiteSettingID) UUID() uuid.UUID { return uuid.UUID(id) }

// String returns the canonical UUID string representation.
func (id SiteSettingID) String() string { return uuid.UUID(id).String() }

// IsZero reports whether the SiteSettingID is the zero value.
func (id SiteSettingID) IsZero() bool { return uuid.UUID(id) == uuid.Nil }

// MarshalJSON serializes the SiteSettingID as a canonical UUID string.
func (id SiteSettingID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON parses a SiteSettingID from a JSON UUID string.
func (id *SiteSettingID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseSiteSettingID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// Value implements driver.Valuer for SQL driver interop.
func (id SiteSettingID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// Scan implements sql.Scanner for SQL driver interop.
func (id *SiteSettingID) Scan(src any) error {
	var u uuid.UUID
	if err := u.Scan(src); err != nil {
		return err
	}
	*id = SiteSettingID(u)
	return nil
}
