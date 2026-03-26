package domain

import (
	"time"

	"github.com/google/uuid"
)

// TemplateUpdated is a domain event emitted when any field of an email template changes.
// Consumers can use this to invalidate cached rendered templates or trigger preview regeneration.
type TemplateUpdated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Name        string
}

func NewTemplateUpdated(id uuid.UUID, name string) TemplateUpdated {
	return TemplateUpdated{
		aggregateID: id,
		occurredAt:  time.Now(),
		Name:        name,
	}
}

func (e TemplateUpdated) EventName() string      { return "emailtemplate.updated" }
func (e TemplateUpdated) OccurredAt() time.Time   { return e.occurredAt }
func (e TemplateUpdated) AggregateID() uuid.UUID  { return e.aggregateID }
