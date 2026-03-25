package domain

import (
	"time"

	"github.com/google/uuid"
)

// TranslationUpdated is raised when a translation is updated.
type TranslationUpdated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewTranslationUpdated(id uuid.UUID) TranslationUpdated {
	return TranslationUpdated{
		aggregateID: id,
		occurredAt:  time.Now(),
	}
}

func (e TranslationUpdated) EventName() string      { return "translation.updated" }
func (e TranslationUpdated) OccurredAt() time.Time   { return e.occurredAt }
func (e TranslationUpdated) AggregateID() uuid.UUID  { return e.aggregateID }
