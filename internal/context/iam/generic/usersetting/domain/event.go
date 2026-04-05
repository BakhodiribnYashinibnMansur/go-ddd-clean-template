package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserSettingChanged is raised on both creation and update of a user setting.
// It carries the full key-value pair so subscribers can update caches or push real-time notifications.
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
func (e UserSettingChanged) OccurredAt() time.Time  { return e.occurredAt }
func (e UserSettingChanged) AggregateID() uuid.UUID { return e.aggregateID }
