package application

import (
	"context"

	"gct/internal/platform/domain"
)

// EventHandler is a function that handles a domain event.
type EventHandler func(ctx context.Context, event domain.DomainEvent) error

// EventBus publishes domain events and allows subscribing to them.
type EventBus interface {
	Publish(ctx context.Context, events ...domain.DomainEvent) error
	Subscribe(eventName string, handler EventHandler) error
}
