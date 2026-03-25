package app

import (
	"gct/internal/shared/application"
)

// RegisterEventSubscribers sets up cross-BC event communication.
func RegisterEventSubscribers(eventBus application.EventBus, bcs *DDDBoundedContexts) {
	// User events → Audit
	// eventBus.Subscribe("user.created", func(ctx, event) { ... })
	// eventBus.Subscribe("user.signed_in", func(ctx, event) { ... })

	// These will be fully implemented when Kafka is set up.
	// For now, this is a placeholder showing the pattern.
	_ = eventBus
	_ = bcs
}
