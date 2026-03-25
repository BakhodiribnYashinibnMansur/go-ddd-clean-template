package domain

import (
	"time"

	"github.com/google/uuid"
)

// WebhookTriggered is raised when a webhook is triggered.
type WebhookTriggered struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	URL         string
}

func NewWebhookTriggered(id uuid.UUID, url string) WebhookTriggered {
	return WebhookTriggered{
		aggregateID: id,
		occurredAt:  time.Now(),
		URL:         url,
	}
}

func (e WebhookTriggered) EventName() string      { return "webhook.triggered" }
func (e WebhookTriggered) OccurredAt() time.Time   { return e.occurredAt }
func (e WebhookTriggered) AggregateID() uuid.UUID  { return e.aggregateID }
