package eventbus

import (
	"context"
	"sync"

	"gct/internal/shared/application"
	"gct/internal/shared/domain"
)

// Compile-time check that InMemoryEventBus implements application.EventBus.
var _ application.EventBus = (*InMemoryEventBus)(nil)

// InMemoryEventBus implements application.EventBus for development and testing.
type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]application.EventHandler
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]application.EventHandler),
	}
}

func (b *InMemoryEventBus) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, event := range events {
		handlers := b.handlers[event.EventName()]
		for _, handler := range handlers {
			if err := handler(ctx, event); err != nil {
				return err // fail fast for in-memory
			}
		}
	}
	return nil
}

func (b *InMemoryEventBus) Subscribe(eventName string, handler application.EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
	return nil
}
