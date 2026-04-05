package app

import (
	"context"

	sessiondomain "gct/internal/context/iam/session/domain"
	"gct/internal/context/iam/user"
	usercommand "gct/internal/context/iam/user/application/command"
	userdomain "gct/internal/context/iam/user/domain"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"
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
			UserID:    userdomain.UserID(e.AggregateID()),
			SessionID: userdomain.SessionID(e.SessionID),
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
			UserID: userdomain.UserID(e.AggregateID()),
		})
	}); err != nil {
		l.Fatalf("failed to subscribe to session.revoke_all_requested: %v", err)
	}
}
