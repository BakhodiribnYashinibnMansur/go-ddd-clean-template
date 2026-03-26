package domain

import (
	"time"

	"github.com/google/uuid"
)

// SettingUpdated is raised when any field of a site setting changes.
// Subscribers should use this to invalidate cached configuration values.
type SettingUpdated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewSettingUpdated(id uuid.UUID) SettingUpdated {
	return SettingUpdated{
		aggregateID: id,
		occurredAt:  time.Now(),
	}
}

func (e SettingUpdated) EventName() string      { return "sitesetting.updated" }
func (e SettingUpdated) OccurredAt() time.Time   { return e.occurredAt }
func (e SettingUpdated) AggregateID() uuid.UUID  { return e.aggregateID }
