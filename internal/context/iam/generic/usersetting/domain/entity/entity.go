package entity

import (
	"time"

	"gct/internal/context/iam/generic/usersetting/domain/event"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// UserSetting is the aggregate root for per-user configuration (e.g., theme, locale, notification preferences).
// Each (userID, key) pair is unique — the repository enforces this via Upsert semantics.
// Values are stored as opaque strings; type interpretation is the caller's responsibility.
type UserSetting struct {
	shared.AggregateRoot
	userID uuid.UUID
	key    string
	value  string
}

// NewUserSetting creates a new UserSetting aggregate and raises a UserSettingChanged event.
func NewUserSetting(userID uuid.UUID, key, value string) *UserSetting {
	us := &UserSetting{
		AggregateRoot: shared.NewAggregateRoot(),
		userID:        userID,
		key:           key,
		value:         value,
	}
	us.AddEvent(event.NewUserSettingChanged(us.ID(), userID, key, value))
	return us
}

// ReconstructUserSetting rebuilds a UserSetting aggregate from persisted data.
func ReconstructUserSetting(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	userID uuid.UUID,
	key, value string,
) *UserSetting {
	return &UserSetting{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, nil),
		userID:        userID,
		key:           key,
		value:         value,
	}
}

// ChangeValue updates the setting value and raises a UserSettingChanged event.
// The event carries the full (userID, key, value) tuple so subscribers can react without re-querying.
func (us *UserSetting) ChangeValue(value string) {
	us.value = value
	us.Touch()
	us.AddEvent(event.NewUserSettingChanged(us.ID(), us.userID, us.key, value))
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (us *UserSetting) TypedID() UserSettingID { return UserSettingID(us.ID()) }
func (us *UserSetting) UserID() uuid.UUID      { return us.userID }
func (us *UserSetting) Key() string            { return us.key }
func (us *UserSetting) Value() string          { return us.value }
