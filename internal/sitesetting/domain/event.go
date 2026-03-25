package domain

import (
	"time"

	"github.com/google/uuid"
)

// SettingUpdated is raised when a site setting is updated.
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
