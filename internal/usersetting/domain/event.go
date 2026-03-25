package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserSettingChanged is raised when a user setting is created or updated.
type UserSettingChanged struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	UserID      uuid.UUID
	Key         string
	Value       string
}

func NewUserSettingChanged(id, userID uuid.UUID, key, value string) UserSettingChanged {
	return UserSettingChanged{
		aggregateID: id,
		occurredAt:  time.Now(),
		UserID:      userID,
		Key:         key,
		Value:       value,
	}
}

func (e UserSettingChanged) EventName() string      { return "usersetting.changed" }
func (e UserSettingChanged) OccurredAt() time.Time   { return e.occurredAt }
func (e UserSettingChanged) AggregateID() uuid.UUID  { return e.aggregateID }
