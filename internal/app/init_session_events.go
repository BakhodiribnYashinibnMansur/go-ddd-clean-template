package app

import (
	"context"

	sessiondomain "gct/internal/session/domain"
	"gct/internal/shared/application"
	shareddomain "gct/internal/shared/domain"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user"
	usercommand "gct/internal/user/application/command"
)

// subscribeSessionEvents wires session domain events to User BC handlers.
// When the Session BC publishes a revoke request, the User BC performs the actual
// revocation on the user aggregate — keeping both BCs decoupled.
func subscribeSessionEvents(eventBus application.EventBus, userBC *user.BoundedContext, l logger.Log) {
	if err := eventBus.Subscribe("session.revoke_requested", func(ctx context.Context, event shareddomain.DomainEvent) error {
		e, ok := event.(sessiondomain.SessionRevokeRequested)
		if !ok {
			return nil
		}

		l.Infoc(ctx, "handling session.revoke_requested",
			"user_id", e.AggregateID(),
			"session_id", e.SessionID,
		)

		return userBC.SignOut.Handle(ctx, usercommand.SignOutCommand{
			UserID:    e.AggregateID(),
			SessionID: e.SessionID,
		})
	}); err != nil {
		l.Fatalf("failed to subscribe to session.revoke_requested: %v", err)
	}

	if err := eventBus.Subscribe("session.revoke_all_requested", func(ctx context.Context, event shareddomain.DomainEvent) error {
		e, ok := event.(sessiondomain.SessionRevokeAllRequested)
		if !ok {
			return nil
		}

		l.Infoc(ctx, "handling session.revoke_all_requested",
			"user_id", e.AggregateID(),
		)

		return userBC.RevokeAll.Handle(ctx, usercommand.RevokeAllSessionsCommand{
			UserID: e.AggregateID(),
		})
	}); err != nil {
		l.Fatalf("failed to subscribe to session.revoke_all_requested: %v", err)
	}
}
