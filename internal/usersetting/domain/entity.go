package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// UserSetting is the aggregate root for user settings (key-value pairs per user).
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
	us.AddEvent(NewUserSettingChanged(us.ID(), userID, key, value))
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

// ChangeValue updates the setting value.
func (us *UserSetting) ChangeValue(value string) {
	us.value = value
	us.Touch()
	us.AddEvent(NewUserSettingChanged(us.ID(), us.userID, us.key, value))
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (us *UserSetting) UserID() uuid.UUID { return us.userID }
func (us *UserSetting) Key() string       { return us.key }
func (us *UserSetting) Value() string     { return us.value }
