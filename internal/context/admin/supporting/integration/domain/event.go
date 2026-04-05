package domain

import (
	"time"

	"github.com/google/uuid"
)

// IntegrationConnected is a domain event raised when a new third-party integration is registered.
// Subscribers may use this to perform initial connectivity checks or provision webhooks.
type IntegrationConnected struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Name        string
	Type        string
}

func NewIntegrationConnected(id uuid.UUID, name, intType string) IntegrationConnected {
	return IntegrationConnected{
		aggregateID: id,
		occurredAt:  time.Now(),
		Name:        name,
		Type:        intType,
	}
}

func (e IntegrationConnected) EventName() string      { return "integration.connected" }
func (e IntegrationConnected) OccurredAt() time.Time  { return e.occurredAt }
func (e IntegrationConnected) AggregateID() uuid.UUID { return e.aggregateID }
